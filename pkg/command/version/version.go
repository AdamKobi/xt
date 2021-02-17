package command

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/adamkobi/xt/pkg/cmdutil"
	"github.com/spf13/cobra"
)

//NewCmdVersion creates a version command
func NewCmdVersion(f *cmdutil.Factory, version, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(f.IOStreams.Out, Format(version, buildDate))
		},
	}

	return cmd
}

func Format(version, buildDate string) string {
	version = strings.TrimPrefix(version, "v")

	var dateStr string
	if buildDate != "" {
		dateStr = fmt.Sprintf(" (%s)", buildDate)
	}

	return fmt.Sprintf("xt version %s%s\n%s\n", version, dateStr, changelogURL(version))
}

func changelogURL(version string) string {
	path := "https://github.com/adamkobi/xt"
	r := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w.]+)?$`)
	if !r.MatchString(version) {
		return fmt.Sprintf("%s/releases/latest", path)
	}

	url := fmt.Sprintf("%s/releases/tag/v%s", path, strings.TrimPrefix(version, "v"))
	return url
}
