package util

const (
	Success         = 0
	ErrorIo         = 1
	ErrorDownload   = 2
	ErrorUnknownCmd = 3
)

type Args struct {
	Help    bool
	Version bool
	Config  string
}

type Release struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}
