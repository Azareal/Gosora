// Work in progress
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	"gopkg.in/olivere/elastic.v6"
)

func main() {
	log.Print("Loading the configuration data")
	err := common.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Processing configuration data")
	err = common.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	if common.DbConfig.Adapter != "mysql" && common.DbConfig.Adapter != "" {
		log.Fatal("Only MySQL is supported for upgrades right now, please wait for a newer build of the patcher")
	}

	err = prepMySQL()
	if err != nil {
		log.Fatal(err)
	}

	client, err := elastic.NewClient(elastic.SetErrorLog(log.New(os.Stdout, "ES ", log.LstdFlags)))
	if err != nil {
		log.Fatal(err)
	}
	_, _, err = client.Ping("http://127.0.0.1:9200").Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	err = setupIndices(client)
	if err != nil {
		log.Fatal(err)
	}

	err = setupData(client)
	if err != nil {
		log.Fatal(err)
	}
}

func prepMySQL() error {
	return qgen.Builder.Init("mysql", map[string]string{
		"host":      common.DbConfig.Host,
		"port":      common.DbConfig.Port,
		"name":      common.DbConfig.Dbname,
		"username":  common.DbConfig.Username,
		"password":  common.DbConfig.Password,
		"collation": "utf8mb4_general_ci",
	})
}

type ESIndexBase struct {
	Mappings ESIndexMappings `json:"mappings"`
}

type ESIndexMappings struct {
	Doc ESIndexDoc `json:"_doc"`
}

type ESIndexDoc struct {
	Properties map[string]map[string]string `json:"properties"`
}

type ESDocMap map[string]map[string]string

func (d ESDocMap) Add(column string, cType string) {
	d["column"] = map[string]string{"type": cType}
}

func setupIndices(client *elastic.Client) error {
	exists, err := client.IndexExists("topics").Do(context.Background())
	if err != nil {
		return err
	}
	if exists {
		deleteIndex, err := client.DeleteIndex("topics").Do(context.Background())
		if err != nil {
			return err
		}
		if !deleteIndex.Acknowledged {
			return errors.New("delete not acknowledged")
		}
	}

	docMap := make(ESDocMap)
	docMap.Add("tid", "integer")
	docMap.Add("title", "text")
	docMap.Add("content", "text")
	docMap.Add("createdBy", "integer")
	docMap.Add("ip", "ip")
	docMap.Add("suggest", "completion")
	indexBase := ESIndexBase{ESIndexMappings{ESIndexDoc{docMap}}}
	oBytes, err := json.Marshal(indexBase)
	if err != nil {
		return err
	}
	createIndex, err := client.CreateIndex("topics").Body(string(oBytes)).Do(context.Background())
	if err != nil {
		return err
	}
	if !createIndex.Acknowledged {
		return errors.New("not acknowledged")
	}

	exists, err = client.IndexExists("replies").Do(context.Background())
	if err != nil {
		return err
	}
	if exists {
		deleteIndex, err := client.DeleteIndex("replies").Do(context.Background())
		if err != nil {
			return err
		}
		if !deleteIndex.Acknowledged {
			return errors.New("delete not acknowledged")
		}
	}

	docMap = make(ESDocMap)
	docMap.Add("rid", "integer")
	docMap.Add("tid", "integer")
	docMap.Add("content", "text")
	docMap.Add("createdBy", "integer")
	docMap.Add("ip", "ip")
	docMap.Add("suggest", "completion")
	indexBase = ESIndexBase{ESIndexMappings{ESIndexDoc{docMap}}}
	oBytes, err = json.Marshal(indexBase)
	if err != nil {
		return err
	}
	createIndex, err = client.CreateIndex("replies").Body(string(oBytes)).Do(context.Background())
	if err != nil {
		return err
	}
	if !createIndex.Acknowledged {
		return errors.New("not acknowledged")
	}

	return nil
}

type ESTopic struct {
	ID        int    `json:"tid"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedBy int    `json:"createdBy"`
	IPAddress string `json:"ip"`
}

type ESReply struct {
	ID        int    `json:"rid"`
	TID       int    `json:"tid"`
	Content   string `json:"content"`
	CreatedBy int    `json:"createdBy"`
	IPAddress string `json:"ip"`
}

func setupData(client *elastic.Client) error {
	err := qgen.NewAcc().Select("topics").Cols("tid, title, content, createdBy, ipaddress").Each(func(rows *sql.Rows) error {
		var tid, createdBy int
		var title, content, ip string
		err := rows.Scan(&tid, &title, &content, &createdBy, &ip)
		if err != nil {
			return err
		}

		topic := ESTopic{tid, title, content, createdBy, ip}
		_, err = client.Index().Index("topics").Type("_doc").Id(strconv.Itoa(tid)).BodyJson(topic).Do(context.Background())
		return err
	})
	if err != nil {
		return err
	}

	return qgen.NewAcc().Select("replies").Cols("rid, tid, content, createdBy, ipaddress").Each(func(rows *sql.Rows) error {
		var rid, tid, createdBy int
		var content, ip string
		err := rows.Scan(&rid, &tid, &content, &createdBy, &ip)
		if err != nil {
			return err
		}

		reply := ESReply{rid, tid, content, createdBy, ip}
		_, err = client.Index().Index("replies").Type("_doc").Id(strconv.Itoa(rid)).BodyJson(reply).Do(context.Background())
		return err
	})
}
