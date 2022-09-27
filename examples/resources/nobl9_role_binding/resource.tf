resource "nobl9_role_binding" "this" {
  name        = "my-role-binding"
  user        = "1234567890asdfghjkl"
  role_ref    = "project-owner"
  project_ref = "1234567890asdfghjkl"
}