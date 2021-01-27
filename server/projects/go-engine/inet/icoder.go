package inet

type IAttachParas interface {
	FillNetStream() []byte
}

type ICoder interface {
	GetHeaderLen() uint32
	GetBodyLen(datas []byte) (uint32, error)

	EnCodeBody(datas []byte) ([]byte, bool)
	DecodeBody(datas []byte) ([]byte, error)

	ZipBody(datas []byte) ([]byte, bool)
	UnzipBody(datas []byte) ([]byte, error)

	FillNetStream(datas []byte) ([]byte, error)
}
