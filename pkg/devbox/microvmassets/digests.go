package microvmassets

// PinnedDigests are the sha256 values of the default microvm SandboxConfig version (v2) shipped by the
// agent-substrate chart (files/sandboxconfig/microvm.yaml), keyed arch -> asset key. Update them together with
// DefaultKataVersion and DefaultCloudHypervisorVersion when the chart bumps its pins, or pass a SandboxConfig.
var PinnedDigests = map[string]map[string]string{
	"arm64": {
		"cloud-hypervisor": "f192b510eea1c710cbc439d716bb0573c223fc463dbe3e6523788a2b7ef62850",
		"virtiofsd":        "4b43f4f6de6500203b01214a790314eea4182f5f110594ddb28ae2805622f06f",
		"kata-kernel":      "dece388830720fe8bddb41552fff5f486240aeece5acc0aebc1812761e3e3f66",
		"kata-image":       "69f9a76e5a2563b7c74d1f271f3a904240f50ef405465f832021525894519f97",
	},
	"amd64": {
		"cloud-hypervisor": "448af3d4e59b22c2987f7df94c213ad40fb53a10d437e42b5ee6c4fce7c29ecc",
		"virtiofsd":        "15b2e72a78cc08a9bd8a6943e89fb69c88cb3cbeb63069efceade835342ac7d4",
		"kata-kernel":      "f87d8359035cb7557b16a319e6431ea09b183e62268aef436bf8dc6530d786b6",
		"kata-image":       "6770db2cca5710593f27c9ac5e1d9d7aab836422695a5e5c4807496a92a0f24d",
	},
}
