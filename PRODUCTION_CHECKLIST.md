# Production Deployment Checklist

## Pre-Deployment

### Environment Variables
- [ ] Update `ololo-backend/.env.production` - DB_PASSWORD
- [ ] Update `ololo-backend/.env.production` - JWT_SECRET
- [ ] Update `ololo-backend/.env.production` - INIT_ADMIN_PASSWORD
- [ ] Update `ololo-backend/.env.production` - CORS_ALLOWED_ORIGINS
- [ ] Update `ololo-backend/.env.production` - THIRD_PARTY_API_URL
- [ ] Update `dummy-backend-api/.env.production` - DB_PASSWORD
- [ ] Verify all environment variables are set (no defaults)

### Docker & System
- [ ] Docker Engine 20.10+ installed
- [ ] Docker Compose 1.29+ installed
- [ ] Sufficient disk space (check: `docker system df`)
- [ ] Required ports available (3000, 8080)
- [ ] Git installed and configured

### Network & Security
- [ ] Firewall rules configured for ports 3000, 8080
- [ ] SSL/TLS certificates ready (for reverse proxy)
- [ ] Reverse proxy configured (nginx/traefik)
- [ ] DNS entries configured
- [ ] IP whitelisting configured (if applicable)

### Database
- [ ] PostgreSQL credentials changed in .env.production
- [ ] Database backups configured
- [ ] Backup test successful
- [ ] Database migration plan ready

## Deployment

### Start Services
```bash
cd backend
./deploy.sh prod rebuild
```

### Verification
- [ ] Both containers running: `docker ps`
- [ ] No critical errors in logs: `docker-compose logs`
- [ ] Dummy API responding: `curl http://localhost:3000`
- [ ] Ololo API responding: `curl http://localhost:8080`

### Health Checks
- [ ] Container health checks passing: `docker-compose ps`
- [ ] Database connections successful
- [ ] Inter-service communication working
- [ ] No restart loops

## Post-Deployment

### Monitoring
- [ ] Logs being collected/monitored
- [ ] Resource usage monitored
- [ ] Error tracking enabled
- [ ] Uptime monitoring configured

### Maintenance
- [ ] Backup schedule configured
- [ ] Log rotation configured
- [ ] Update notifications set up
- [ ] Rollback plan documented

### Testing
- [ ] API endpoints tested
- [ ] Database operations tested
- [ ] External service integrations tested
- [ ] Load testing completed (if applicable)

## Ongoing Operations

### Regular Tasks
- [ ] Weekly backup verification
- [ ] Monthly security updates review
- [ ] Monthly log review
- [ ] Quarterly capacity planning

### Documentation
- [ ] Deployment process documented
- [ ] Troubleshooting guide created
- [ ] Rollback procedure documented
- [ ] Team trained on operations

## Emergency Contacts & Escalation

- [ ] On-call contact configured
- [ ] Escalation procedure documented
- [ ] Critical issue response time defined
- [ ] Incident log setup

## Sign-Off

- [ ] Production lead approval: _________________
- [ ] Security review: _________________
- [ ] Date deployed: _________________
- [ ] Deployment notes: _________________________________________

---

**Quick Commands for Production:**

```bash
# Start services
./deploy.sh prod rebuild

# Monitor
docker-compose logs -f

# Stop (if needed)
docker-compose down

# Database backup
docker exec ololo-postgres-prod pg_dump -U postgres -d ololo_gate > backup.sql

# Restart specific service
docker-compose restart dummy-api
```

**Critical Environment Variables to Update:**

```env
# ololo-backend/.env.production
DB_PASSWORD=SECURE_PASSWORD_HERE
JWT_SECRET=SECURE_JWT_SECRET_HERE
INIT_ADMIN_PASSWORD=SECURE_ADMIN_PASSWORD_HERE
CORS_ALLOWED_ORIGINS=https://yourdomain.com
THIRD_PARTY_API_URL=https://api.yourdomain.com:3000

# dummy-backend-api/.env.production
DB_PASSWORD=SECURE_PASSWORD_HERE
```
