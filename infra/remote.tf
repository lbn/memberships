terraform {
  backend "remote" {
    workspaces {
      name = "memberships"
    }
  }
}
