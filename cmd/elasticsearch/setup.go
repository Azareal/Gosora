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

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	"gopkg.in/olivere/elastic.v6"
)

func main() {
	log.Print("Loading the configuration data")
	err := c.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Processing configuration data")
	err = c.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	if c.DbConfig.Adapter != "mysql" && c.DbConfig.Adapter != "" {
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
		"host":      c.DbConfig.Host,
		"port":      c.DbConfig.Port,
		"name":      c.DbConfig.Dbname,
		"username":  c.DbConfig.Username,
		"password":  c.DbConfig.Password,
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
	IP string `json:"ip"`
}

type ESReply struct {
	ID        int    `json:"rid"`
	TID       int    `json:"tid"`
	Content   string `json:"content"`
	CreatedBy int    `json:"createdBy"`
	IP string `json:"ip"`
}

func setupData(client *elastic.Client) error {
	tcount := 4
	errs := make(chan error)

	go func() {
		tin := make([]chan ESTopic, tcount)
		tf := func(tin chan ESTopic) {
			for {
				topic, more := <-tin
				if !more {
					break
				}
				_, err := client.Index().Index("topics").Type("_doc").Id(strconv.Itoa(topic.ID)).BodyJson(topic).Do(context.Background())
				if err != nil {
					errs <- err
				}
			}
		}
		for i := 0; i < 4; i++ {
			go tf(tin[i])
		}

		oi := 0
		err := qgen.NewAcc().Select("topics").Cols("tid, title, content, createdBy, ipaddress").Each(func(rows *sql.Rows) error {
			t := ESTopic{}
			err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CreatedBy, &t.IP)
			if err != nil {
				return err
			}
			tin[oi] <- t
			if oi < 3 {
				oi++
			}
			return nil
		})
		for i := 0; i < 4; i++ {
			close(tin[i])
		}
		errs <- err
	}()

	go func() {
		rin := make([]chan ESReply, tcount)
		rf := func(rin chan ESReply) {
			for {
				reply, more := <-rin
				if !more {
					break
				}
				_, err := client.Index().Index("replies").Type("_doc").Id(strconv.Itoa(reply.ID)).BodyJson(reply).Do(context.Background())
				if err != nil {
					errs <- err
				}
			}
		}
		for i := 0; i < 4; i++ {
			rf(rin[i])
		}
		oi := 0
		err := qgen.NewAcc().Select("replies").Cols("rid, tid, content, createdBy, ipaddress").Each(func(rows *sql.Rows) error {
			r := ESReply{}
			err := rows.Scan(&r.ID, &r.TID, &r.Content, &r.CreatedBy, &r.IP)
			if err != nil {
				return err
			}
			rin[oi] <- r
			if oi < 3 {
				oi++
			}
			return nil
		})
		for i := 0; i < 4; i++ {
			close(rin[i])
		}
		errs <- err
	}()

	fin := 0
	for {
		err := <-errs
		if err == nil {
			fin++
			if fin == 2 {
				return nil
			}
		} else {
			return err
		}
	}
}
