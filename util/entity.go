package util

const (
	Success         = 0
	Error           = 1
	ErrorIO         = 2
	ErrorDownload   = 3
	ErrorUnknownCmd = 4
)

const (
	SyncLatest        = "latest"
	SyncLatestRelease = "latest_release"
	SyncAll           = "all"
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
