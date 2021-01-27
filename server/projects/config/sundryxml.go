package config

type SundryXml struct {
}

func ReadSundryXml(path string) error {

	return nil
}

var GSundryXml *SundryXml

func init() {
	GSundryXml = &SundryXml{}
}
