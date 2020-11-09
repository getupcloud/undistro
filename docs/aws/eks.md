# EKS Support

To create EKS cluster just set `managed: true` in infrastructureProvider section of cluster spec.

```yaml
  infrastructureProvider:
    name: aws
    managed: true
    sshKey: <YOUR SSH KEY NAME>
    env:
      - name: AWS_ACCESS_KEY_ID
        value: ""
      - name: AWS_SECRET_ACCESS_KEY
        value: ""
      - name: AWS_REGION
        value: ""
```
```