package nobl9

import "fmt"

func mockAlertMethod(name, project string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_slack" "%s" {
  name        = "%s"
  project     = "%s"
  description = "slack"
  url         = "https://slack.com"
}
`, name, name, project)
}
