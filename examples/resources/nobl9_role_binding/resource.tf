# Preferred: using account_id
resource "nobl9_role_binding" "this" {
  name        = "my-role-binding"
  account_id  = "accountID1234"
  role_ref    = "project-owner"
  project_ref = "default"
}

# Group-based role binding
resource "nobl9_role_binding" "group_binding" {
  name        = "group-role-binding"
  group_ref   = "test"
  role_ref    = "project-owner"
  project_ref = "default"
}

# Deprecated: using user field (backward compatibility)
resource "nobl9_role_binding" "legacy" {
  name        = "legacy-role-binding"
  user        = "accountID1234"  # Deprecated: use account_id instead
  role_ref    = "project-owner"
  project_ref = "default"
}
