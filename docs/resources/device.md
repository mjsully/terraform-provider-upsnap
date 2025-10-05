# Device
 
## Example Usage
 
```hcl
// An example of how to use this resource
resource "upsnap_device" "test_device" {
  name = "test_device"
  ip = "192.168.1.100"
  mac = "AA:BB:CC:DD:EE:FF"
  netmask = "255.255.255.0"
  description = "A test device to be added to UpSnap."
  link = "https://www.google.co.uk"
  groups = [upsnap_device_group.test_group.id]
}
```
 
## Argument Reference
 
* `name` - (Required) Name of the device.
* `ip` - (Required) IP address of the device.
* `mac` - (Required) Mac address of the device.
* `netmask` - (Required) Netmask of the device.
* `description` - (Optional) Description.
* `link` - (Optional) Enable clicking on the device name in the UI to open a link.
* `groups` - (Optional) A list of device group IDs that this device should be in.