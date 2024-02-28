resource "nobl9_role_binding" "this" {
  name        = "my-role-binding"
  user        = "test"
  role_ref    = "project-owner"
  project_ref = "default"
}

resource "nobl9_role_binding" "this" {
  name        = "group-role-binding"
  group_ref   = "test"
  role_ref    = "project-owner"
  project_ref = "default"
}
