package cmd

type Options struct {
	Password     string       `kong:"short='p',help='Password when want login'"`
	LoginID      string       `kong:"short='l',help='Login as other ID'"`
	Verbose      bool         `kong:"short='v',help='Enable verbose mode'"`
	Debug        bool         `kong:"hidden,help='Print debug info'"`
	PrintVersion PrintVersion `kong:"hidden,help='Print version'"`

	Logger Logger `kong:"-"`
}

//nolint:maligned
type CLI struct {
	Options
	Dump    Dump    `kong:"cmd,help='Download all'"`
	Version Version `kong:"cmd,help='Print version'"`
}
