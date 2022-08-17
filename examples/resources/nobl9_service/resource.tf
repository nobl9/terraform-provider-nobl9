resource "nobl9_project" "this" {
  display_name = "Foo Project"
  name         = "foo-project"
  description  = "An example terraform project"
}

resource "nobl9_service" "this" {
  name         = "${nobl9_project.this.name}-front-page"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Front Page"
  description  = "Front page service"
}
