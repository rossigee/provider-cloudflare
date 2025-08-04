# Complete Website Integration Example

This example demonstrates a comprehensive, production-ready website setup using Cloudflare provider resources.

## 🎯 Scenario

Setting up a complete web infrastructure for `example.com` with:
- Modern DNS configuration with IPv4/IPv6 support
- Geographic load balancing with health monitoring
- Advanced security protection (WAF, rate limiting, geo-blocking)
- Performance optimization (caching, compression)
- Email delivery configuration (MX, SPF, DMARC)
- SRV records for additional services

## 📋 Resources Created

### 1. Provider Configuration (`00-provider-config.yaml`)
- Kubernetes secret with Cloudflare API credentials
- ProviderConfig for authentication

### 2. DNS Zone (`01-zone.yaml`)
- Zone with comprehensive security and performance settings
- SSL/TLS configuration (Flexible SSL, HTTPS redirect)
- Modern protocols (HTTP/2, HTTP/3, Zero RTT)
- Development and production optimizations

### 3. DNS Records (`02-dns-records.yaml`)
- **A/AAAA Records**: Root domain and API subdomain with proxy
- **CNAME**: WWW subdomain pointing to root
- **MX Record**: Email delivery configuration
- **TXT Records**: SPF and DMARC for email authentication
- **SRV Record**: SIP service configuration

### 4. Load Balancing (`03-load-balancing.yaml`)
- **Health Monitor**: HTTPS health checks with custom path
- **Primary Pool**: Main web servers with round-robin
- **Backup Pool**: Failover servers with least-connections
- **Load Balancer**: Geographic routing with session affinity

### 5. Security Rules (`04-security-rules.yaml`)
- **Firewall Rules**: Block malicious IPs and attack patterns
- **Rate Limiting**: API endpoint protection
- **Bot Management**: Challenge suspicious bots, allow search engines
- **Admin Protection**: Enhanced security for admin areas
- **Transform Rules**: URL redirects and security headers

### 6. Cache Rules (`05-cache-rules.yaml`)
- **Static Assets**: Long-term caching for CSS, JS, images
- **API Responses**: Short-term caching for public APIs
- **HTML Pages**: Medium-term caching with device detection
- **Admin Bypass**: No caching for authenticated areas
- **Dynamic Content**: Very short caching for real-time data

## 🚀 Deployment Order

Deploy resources in this specific order to respect dependencies:

```bash
# 1. Configure provider authentication
kubectl apply -f 00-provider-config.yaml

# 2. Create the DNS zone (required for all other resources)
kubectl apply -f 01-zone.yaml

# 3. Wait for zone to be ready, then create DNS records
kubectl wait --for=condition=Ready zone/example-website-zone --timeout=300s
kubectl apply -f 02-dns-records.yaml

# 4. Set up load balancing (optional but recommended)
kubectl apply -f 03-load-balancing.yaml

# 5. Configure security rules
kubectl apply -f 04-security-rules.yaml

# 6. Optimize with cache rules  
kubectl apply -f 05-cache-rules.yaml
```

## ⚙️ Customization

Before deploying, customize these values:

### Required Changes
- Replace `example.com` with your actual domain name
- Update IP addresses (`192.0.2.x`) with your actual server IPs
- Replace IPv6 addresses (`2001:db8::x`) with your actual IPv6 addresses
- Update `your-cloudflare-api-token-here` with your actual API token

### Optional Customizations
- Adjust TTL values based on your change frequency
- Modify cache rules for your specific content types
- Update security rules for your threat model
- Configure additional geographic regions in load balancer
- Add more health check endpoints

## 🔍 Validation

After deployment, verify the setup:

```bash
# Check resource status
kubectl get zones,records,loadbalancers,rulesets,cacherules

# Verify DNS propagation
nslookup example.com
nslookup www.example.com

# Test load balancer
curl -H "Host: lb.example.com" http://your-load-balancer-ip/health

# Check security rules (should be challenged/blocked)
curl -H "User-Agent: malicious-bot" https://example.com/
```

## 🛡️ Security Considerations

- **API Token Permissions**: Use least-privilege tokens with only required permissions
- **IP Allowlisting**: Consider restricting access to admin areas by IP
- **Rate Limiting**: Adjust rate limits based on your traffic patterns
- **Geo-blocking**: Enable country blocking if appropriate for your use case
- **SSL Settings**: Consider upgrading to "Full (strict)" SSL mode in production

## 📊 Monitoring

Monitor your deployment:

```bash
# Check resource conditions
kubectl describe zone example-website-zone
kubectl describe loadbalancer website-load-balancer

# View connection details
kubectl get secret zone-connection-details -o yaml

# Monitor events
kubectl get events --field-selector involvedObject.kind=Zone
```

## 🔄 Cleanup

To remove all resources:

```bash
# Remove in reverse order
kubectl delete -f 05-cache-rules.yaml
kubectl delete -f 04-security-rules.yaml  
kubectl delete -f 03-load-balancing.yaml
kubectl delete -f 02-dns-records.yaml
kubectl delete -f 01-zone.yaml
kubectl delete -f 00-provider-config.yaml
```

## 🚨 Production Notes

- Test all configurations in a development zone first
- Plan maintenance windows for DNS changes
- Keep backups of your Cloudflare configurations
- Monitor traffic patterns after deploying cache rules
- Review security logs regularly for blocked threats
- Update health check endpoints as your infrastructure evolves