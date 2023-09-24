package common

var File FileData

type FileData struct {
	Toml []byte // Default config file
}

func InitFiles(toml []byte) {

	// Init embedded files
	File = FileData{
		Toml: toml,
	}
}
