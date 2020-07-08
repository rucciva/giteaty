package hydrahandler

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"

	"github.com/gtank/cryptopasta"
	"github.com/ory/hydra-client-go/models"
	"golang.org/x/oauth2"
)

type loginContext struct {
	Token *oauth2.Token `json:"token"`
}

func loginContextFromRaw(raw models.JSONRawMessage) (lc loginContext) {
	b, _ := json.Marshal(raw)
	_ = json.Unmarshal(b, &lc)
	return
}

// nolint:deadcode
func loginContextFromChipper(str string, key *[32]byte) (lc loginContext, err error) {
	b, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return
	}
	plain, err := cryptopasta.Decrypt(b, key)
	if err != nil {
		return
	}
	err = gob.NewDecoder(bytes.NewReader(plain)).Decode(&lc)
	return
}

func (lc loginContext) encrypt(key *[32]byte) (str string, err error) {
	var buf bytes.Buffer
	if err = gob.NewEncoder(&buf).Encode(lc); err != nil {
		return
	}
	b, err := cryptopasta.Encrypt(buf.Bytes(), key)
	if err != nil {
		return
	}
	str = base64.RawStdEncoding.EncodeToString(b)
	return
}
