package signchecker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/alphaonly/gomart/internal/schema"
)

type Signer interface {
	IsValidSign(mj schema.MetricsJSON) bool
	Hash(mj schema.MetricsJSON) (hash string, err error)
	Sign(mj *schema.MetricsJSON) (err error)
}

type CheckerSHA256 struct {
	key []byte
}

func NewSHA256(key string) Signer {
	return CheckerSHA256{key: []byte(key)}
}
func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func (c CheckerSHA256) IsValidSign(mj schema.MetricsJSON) (result bool) {

	if c.key == nil || len(c.key) == 0 {
		return true
	}
	var leftHash []byte
	var err error

	//mj hash came to confirm the sign
	leftHash, err = hex.DecodeString(mj.Hash)
	logFatal(err)
	//signing a json copy to compare hashes mjCopy and mj
	mjCopy := mj
	err = c.Sign(&mjCopy)
	logFatal(err)

	//get updated hash to compare with left(original)
	rightHash, err := hex.DecodeString(mjCopy.Hash)
	logFatal(err)
	//compare
	ans := hmac.Equal(leftHash, rightHash)
	if !ans {
		log.Printf("KEY:%v", string(c.key))
		var (
			v  float64
			d  int64
			vc float64
			dc int64
		)
		if mj.Value != nil {
			v = *mj.Value
		}
		if mj.Delta != nil {
			d = *mj.Delta
		}
		if mjCopy.Value != nil {
			vc = *mjCopy.Value
		}
		if mjCopy.Delta != nil {
			dc = *mjCopy.Delta
		}

		log.Printf("inbound data: id:%v,type:%v,delta:%v,value:%v,hash:%v", mj.ID, mj.MType, d, v, mj.Hash)
		log.Printf("calc    data: id:%v,type:%v,delta:%v,value:%v,hash:%v", mjCopy.ID, mjCopy.MType, dc, vc, mjCopy.Hash)
		log.Printf("inbound hash:%v", leftHash)
		log.Printf("calcula hash:%v", rightHash)
	}

	return ans
}

var counterHashMessage = func(id string, delta *int64) []byte {
	return []byte(fmt.Sprintf("%s:counter:%d", id, *delta))
}

var gaugeHashMessage = func(id string, value *float64) []byte {
	return []byte(fmt.Sprintf("%s:gauge:%f", id, *value))
}

func (c CheckerSHA256) counterHash(id string, delta *int64) ([]byte, error) {
	msg := counterHashMessage(id, delta)
	h := hmac.New(sha256.New, c.key)
	_, err := h.Write(msg)
	if err != nil {
		return nil, err
	}
	return h.Sum(*new([]byte)), nil
}
func (c CheckerSHA256) gaugeHash(id string, value *float64) ([]byte, error) {
	msg := gaugeHashMessage(id, value)
	h := hmac.New(sha256.New, c.key)
	_, err := h.Write(msg)
	if err != nil {
		return nil, err
	}

	return h.Sum(*new([]byte)), nil
}

func (c CheckerSHA256) Hash(mj schema.MetricsJSON) (hash string, err error) {
	err = c.Sign(&mj)
	if err != nil {
		return "", err
	}
	return mj.Hash, nil
}
func (c CheckerSHA256) Sign(mj *schema.MetricsJSON) (err error) {
	if c.key == nil || len(c.key) == 0 {
		return nil
	}
	var hashBytes []byte

	switch mj.MType {
	case "gauge":
		{
			if mj.Value == nil {
				return errors.New("signer:value is nil")
			}
			hashBytes, err = c.gaugeHash(mj.ID, mj.Value)
		}
	case "counter":
		{
			if mj.Delta == nil {
				return errors.New("signer:delta is  nil")
			}
			hashBytes, err = c.counterHash(mj.ID, mj.Delta)
		}
	default:
		log.Panic("CheckJSON unknown type")
		return
	}
	mj.Hash = hex.EncodeToString(hashBytes)

	if mj.Delta != nil {
		log.Printf("agent: outbound data: key:%v, id:%v,type:%v,delta:%v,hash:%v", string(c.key), mj.ID, mj.MType, *mj.Delta, mj.Hash)
	} else {
		log.Printf("agent: outbound data: key:%v, id:%v,type:%v,value:%v,hash:%v", string(c.key), mj.ID, mj.MType, *mj.Value, mj.Hash)
	}

	return err
}
