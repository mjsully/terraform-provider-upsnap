# UpSnap Provider
 
UpSnap is an awesome wake-on-lan tool, and this provider enables the management of devices in UpSnap through Terraform.
 
## Example Usage
 
```hcl
// UpSnap provider configuration
provider "upsnap" {
  host = <UpSnap host>
  username = <UpSnap username>
  password = <UpSnap password>
}
```