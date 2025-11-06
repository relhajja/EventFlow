up:
	@echo "🚀 Starting all services..."
	@sudo docker compose up -d --build
	@sleep 5
	@sudo docker compose ps
	@echo "✅ Running! Run 'make test' to verify"

down:
	@sudo docker compose down

test:
	@echo "🧪 Testing endpoints..."
	@echo "\n📡 Backend API:"
	@curl -s http://localhost:8080/
	@echo "\n\n🗄️  Database:"
	@sudo docker exec riadcloud-postgres psql -U riadcloud -d riadcloud -c "SELECT COUNT(*) FROM services;"
	@echo "\n🎨 Frontend: http://localhost:3000"
	@echo "✅ All services tested!"

clean:
	@sudo docker compose down -v
	@sudo docker rmi webapp-webapp 2>/dev/null || true
	@sudo docker rmi webapp-frontend 2>/dev/null || true
	@echo "✅ Cleaned"

help:
	@echo "RiadCloud Docker Test"
	@echo "Commands: up, down, test, clean"
