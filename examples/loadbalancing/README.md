# Load Balancing Examples

This directory contains examples for Cloudflare Load Balancing resources.

## Components

Cloudflare Load Balancing consists of three main components:

1. **LoadBalancerMonitor** - Health checks for origin servers
2. **LoadBalancerPool** - Groups of origin servers  
3. **LoadBalancer** - The DNS-based load balancer that routes traffic

## Basic Usage

1. Create a monitor for health checking origins
2. Create one or more pools with origin servers
3. Create a load balancer that uses those pools

## Files

- `monitor.yaml` - Basic HTTP health check monitor
- `pool.yaml` - Basic pool with two origin servers
- `load-balancer.yaml` - Simple geographic load balancer
- `advanced-load-balancer.yaml` - Advanced load balancer with rules and steering
- `full-example.yaml` - Complete example with all components

## Features Demonstrated

- Health monitoring of origin servers
- Geographic traffic steering
- Failover pools
- Session affinity
- Traffic steering rules
- Random and proximity-based routing