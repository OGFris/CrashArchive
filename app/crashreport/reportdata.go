package crashreport

import (
	"fmt"
	"strings"
)

func (rd ReportData) IsValid() error {
	if rd.General.Name != "PocketMine-MP" {
		return fmt.Errorf("spoon detected")
	}

	if rd.General.GIT == strings.Repeat("00", 20) || strings.HasSuffix(rd.General.GIT, "-dirty") {
		return fmt.Errorf("invalid git hash %s in report", rd.General.GIT)
	}
	return nil
}
