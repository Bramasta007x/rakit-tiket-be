package entity

import "rakit-tiket-be/pkg/util"

type Hash string

// func MakeHash(values ...string) Hash {
// 	return Hash(util.MakeHash(true, values...))
// }

func (e Hash) String() string {
	return string(e)
}

func (e Hash) Bytes() []byte {
	return []byte(e.String())
}

func (e Hash) EncodeToBase64() string {
	return util.EncodeToBase64(e.Bytes())
}
