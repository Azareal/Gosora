package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

var ConvoPostProcess ConvoPostProcessor = NewDefaultConvoPostProcessor()

type ConvoPostProcessor interface {
	OnLoad(co *ConversationPost) (*ConversationPost, error)
	OnSave(co *ConversationPost) (*ConversationPost, error)
}

type DefaultConvoPostProcessor struct {
}

func NewDefaultConvoPostProcessor() *DefaultConvoPostProcessor {
	return &DefaultConvoPostProcessor{}
}

func (pr *DefaultConvoPostProcessor) OnLoad(co *ConversationPost) (*ConversationPost, error) {
	return co, nil
}

func (pr *DefaultConvoPostProcessor) OnSave(co *ConversationPost) (*ConversationPost, error) {
	return co, nil
}

type AesConvoPostProcessor struct {
}

func NewAesConvoPostProcessor() *AesConvoPostProcessor {
	return &AesConvoPostProcessor{}
}

func (pr *AesConvoPostProcessor) OnLoad(co *ConversationPost) (*ConversationPost, error) {
	if co.Post != "aes" {
		return co, nil
	}
	key, _ := hex.DecodeString(Config.ConvoKey)

	ciphertext, err := hex.DecodeString(co.Body)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	lco := *co
	lco.Body = string(plaintext)
	return &lco, nil
}

func (pr *AesConvoPostProcessor) OnSave(co *ConversationPost) (*ConversationPost, error) {
	key, _ := hex.DecodeString(Config.ConvoKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(co.Body), nil)

	lco := *co
	lco.Body = hex.EncodeToString(ciphertext)
	lco.Post = "aes"
	return &lco, nil
}

type ConversationPost struct {
	ID        int
	CID       int
	Body      string
	Post      string // aes, ''
	CreatedBy int
}

// TODO: Should we run OnLoad on this? Or maybe add a FetchMeta method to avoid having to decode the message when it's not necessary?
func (co *ConversationPost) Fetch() error {
	return convoStmts.fetchPost.QueryRow(co.ID).Scan(&co.CID, &co.Body, &co.Post, &co.CreatedBy)
}

func (co *ConversationPost) Update() error {
	lco, err := ConvoPostProcess.OnSave(co)
	if err != nil {
		return err
	}
	//GetHookTable().VhookNoRet("convo_post_update", lco)
	_, err = convoStmts.editPost.Exec(lco.Body, lco.Post, lco.ID)
	return err
}

func (co *ConversationPost) Create() (int, error) {
	lco, err := ConvoPostProcess.OnSave(co)
	if err != nil {
		return 0, err
	}
	//GetHookTable().VhookNoRet("convo_post_create", lco)
	res, err := convoStmts.createPost.Exec(lco.CID, lco.Body, lco.Post, lco.CreatedBy)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func (co *ConversationPost) Delete() error {
	_, err := convoStmts.deletePost.Exec(co.ID)
	return err
}
