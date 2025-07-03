# Production Deployment Guide

This guide will help you deploy your social network application to production with proper security and configuration.

## 🚀 Quick Start

1. **Copy environment template**:
   ```bash
   cp .env.production.template .env.production
   ```

2. **Configure your domain and settings** in `.env.production`

3. **Run deployment script**:
   ```bash
   ./deploy-production.sh
   ```

## 📋 Prerequisites

- Docker and Docker Compose installed
- Domain name configured to point to your server
- SSL certificates (Let's Encrypt recommended)
- Server with sufficient resources (minimum 2GB RAM, 2 CPU cores)

## 🔧 Configuration

### Environment Variables

Edit `.env.production` with your production values:

```bash
# Domain Configuration
DOMAIN=your-domain.com
FRONTEND_API_URL=https://your-domain.com/api
FRONTEND_SOCKET_URL=wss://your-domain.com/ws

# Security
AUTH_SECRET_KEY=your-very-secure-random-secret-key-here-at-least-32-characters
ALLOWED_ORIGINS=https://your-domain.com,https://www.your-domain.com
SECURE_COOKIES=true
```

### SSL Certificates

1. **Using Let's Encrypt (Recommended)**:
   ```bash
   # Install certbot
   sudo apt install certbot

   # Get certificates
   sudo certbot certonly --standalone -d your-domain.com -d www.your-domain.com

   # Copy certificates to ssl directory
   sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ./ssl/cert.pem
   sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ./ssl/key.pem
   sudo chown $USER:$USER ./ssl/*.pem
   ```

2. **Using custom certificates**:
   - Place your certificate as `./ssl/cert.pem`
   - Place your private key as `./ssl/key.pem`

### Nginx Configuration

1. Copy the nginx template:
   ```bash
   cp nginx.conf.template nginx.conf
   ```

2. Replace `your-domain.com` with your actual domain in `nginx.conf`

## 🐳 Docker Deployment

### Option 1: With Nginx (Recommended)

```bash
# Deploy with nginx reverse proxy
docker-compose -f docker-compose.prod.yml --env-file .env.production --profile with-nginx up -d
```

### Option 2: Without Nginx

```bash
# Deploy without nginx (you'll need external reverse proxy)
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
```

## 🔒 Security Considerations

### 1. Environment Variables
- **Never commit `.env.production`** to version control
- Use strong, unique values for `AUTH_SECRET_KEY` (minimum 32 characters)
- Restrict `ALLOWED_ORIGINS` to your actual domains only

### 2. SSL/TLS
- Always use HTTPS in production
- Set `SECURE_COOKIES=true`
- Configure proper SSL certificates

### 3. Database Security
- The SQLite database is stored in `./data/` directory
- Ensure proper file permissions and backups
- Consider using external database for high-traffic applications

### 4. File Uploads
- Uploaded files are stored in `./uploads/` directory
- Implement file size limits and type validation
- Consider using cloud storage for scalability

## 📊 Monitoring and Maintenance

### View Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f frontend
docker-compose -f docker-compose.prod.yml logs -f backend
```

### Health Checks
```bash
# Check service status
docker-compose -f docker-compose.prod.yml ps

# Check application health
curl -f https://your-domain.com/api/health || echo "Backend not responding"
```

### Backup Database
```bash
# Create backup
cp ./data/social_network.db ./backups/social_network_$(date +%Y%m%d_%H%M%S).db

# Automated backup script (add to crontab)
0 2 * * * /path/to/your/app/backup-db.sh
```

## 🔄 Updates and Maintenance

### Update Application
```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
docker-compose -f docker-compose.prod.yml --env-file .env.production down
docker-compose -f docker-compose.prod.yml --env-file .env.production build --no-cache
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
```

### Certificate Renewal (Let's Encrypt)
```bash
# Renew certificates
sudo certbot renew

# Update certificates in ssl directory
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ./ssl/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ./ssl/key.pem

# Restart nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

## 🚨 Troubleshooting

### Common Issues

1. **Services won't start**:
   - Check environment variables in `.env.production`
   - Verify Docker and Docker Compose versions
   - Check available disk space and memory

2. **SSL certificate errors**:
   - Verify certificate files exist and have correct permissions
   - Check domain DNS configuration
   - Ensure certificates are not expired

3. **CORS errors**:
   - Verify `ALLOWED_ORIGINS` includes your domain
   - Check that frontend URLs match your domain

4. **WebSocket connection issues**:
   - Ensure WebSocket endpoint is accessible
   - Check nginx WebSocket proxy configuration
   - Verify SSL certificates for WSS connections

### Getting Help

- Check application logs: `docker-compose -f docker-compose.prod.yml logs`
- Verify environment configuration
- Test individual services: `curl -f https://your-domain.com/api/health`

## 📈 Performance Optimization

### For High Traffic
- Consider using external database (PostgreSQL/MySQL)
- Implement Redis for session storage
- Use CDN for static assets
- Set up load balancing with multiple instances
- Monitor resource usage and scale accordingly

### Resource Limits
Add resource limits to `docker-compose.prod.yml`:
```yaml
services:
  frontend:
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

## 🎯 Next Steps

After successful deployment:
1. Set up monitoring (Prometheus/Grafana)
2. Configure automated backups
3. Set up log aggregation
4. Implement health checks
5. Configure alerts for downtime
6. Set up CI/CD pipeline for automated deployments
