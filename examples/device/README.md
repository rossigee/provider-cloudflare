# Device Management Examples

This directory contains examples for managing Cloudflare Device Posture Rules through Crossplane.

## Device Posture Rule

The `device-posture-rule.yaml` file demonstrates how to create device posture rules for endpoint security and compliance checking.

### Key Features Demonstrated

- **Endpoint Security**: Enforce device compliance requirements
- **OS Version Control**: Ensure devices run approved operating system versions
- **Scheduled Checks**: Configure posture evaluation frequency
- **Platform Targeting**: Apply rules to specific operating systems
- **Compliance Monitoring**: Track device posture across your organization

### Usage

```bash
kubectl apply -f device-posture-rule.yaml
```

### Prerequisites

- Cloudflare account with Device Posture capability
- WARP client deployed on managed devices
- Device enrollment configured in Cloudflare Zero Trust

### Device Posture Types

- **os_version**: Check operating system version compliance
- **domain_joined**: Verify domain membership
- **disk_encryption**: Ensure disk encryption is enabled
- **firewall**: Check firewall status
- **client_certificate**: Validate client certificates
- **workspace_one**: Integration with VMware Workspace ONE
- **crowdstrike**: Integration with CrowdStrike Falcon
- **intune**: Integration with Microsoft Intune
- **kolide**: Integration with Kolide device management
- **sentinelone**: Integration with SentinelOne
- **tanium**: Integration with Tanium

### Usage in Access Policies

Device posture rules are typically used in Access policies to control application access based on device compliance:

```yaml
# Example Access Policy using device posture
apiVersion: access.cloudflare.m.crossplane.io/v1beta1
kind: AccessPolicy
metadata:
  name: compliant-devices-only
spec:
  forProvider:
    applicationId: "your-app-id"
    name: "Compliant Devices Policy"
    decision: "allow"
    include:
      - devicePosture: "your-posture-rule-id"
```

### Customization

- Replace `your-account-id` with your Cloudflare account ID
- Choose appropriate `type` based on your compliance requirements
- Adjust `schedule` for check frequency (e.g., "5m", "1h", "24h")
- Set `expiration` for how long posture results are valid
- Configure platform-specific matching criteria

### Monitoring and Reporting

Device posture results can be viewed in:
- Cloudflare Zero Trust dashboard
- Device posture logs
- Access policy evaluation results

### Best Practices

- **Gradual Rollout**: Start with monitoring-only rules before enforcing
- **Clear Naming**: Use descriptive names for easy identification
- **Regular Updates**: Keep OS version requirements current
- **Multiple Rules**: Create separate rules for different device types/groups
- **Testing**: Validate rules with test devices before production deployment