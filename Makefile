.PHONY: build clean install test generate deploy help

# Variables
BINARY_NAME=teleflix
VERSION?=latest
OUTPUT_DIR=./bin
MANIFESTS_DIR=./manifests

# Couleurs pour les messages
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Affiche cette aide
	@echo "$(GREEN)Teleflix - Générateur de manifests Kubernetes$(NC)"
	@echo ""
	@echo "$(YELLOW)Commandes disponibles:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile le binaire
	@echo "$(GREEN)🔨 Compilation du projet...$(NC)"
	@mkdir -p $(OUTPUT_DIR)
	@go build -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "$(GREEN)✅ Binaire compilé: $(OUTPUT_DIR)/$(BINARY_NAME)$(NC)"

clean: ## Nettoie les fichiers générés
	@echo "$(YELLOW)🧹 Nettoyage...$(NC)"
	@rm -rf $(OUTPUT_DIR)
	@rm -rf $(MANIFESTS_DIR)
	@echo "$(GREEN)✅ Nettoyage terminé$(NC)"

install: build ## Installe le binaire dans $GOPATH/bin
	@echo "$(GREEN)📦 Installation...$(NC)"
	@go install ./cmd/main.go
	@echo "$(GREEN)✅ Installation terminée$(NC)"

test: ## Lance les tests
	@echo "$(GREEN)🧪 Lancement des tests...$(NC)"
	@go test -v ./...

generate: build ## Génère les manifests Kubernetes
	@echo "$(GREEN)🚀 Génération des manifests...$(NC)"
	@$(OUTPUT_DIR)/$(BINARY_NAME) --output $(MANIFESTS_DIR)
	@echo "$(GREEN)✅ Manifests générés dans $(MANIFESTS_DIR)$(NC)"

generate-custom: build ## Génère avec une config personnalisée (CONFIG=fichier.yaml)
	@echo "$(GREEN)🚀 Génération avec config personnalisée...$(NC)"
	@$(OUTPUT_DIR)/$(BINARY_NAME) --config $(CONFIG) --output $(MANIFESTS_DIR)

deploy: generate ## Déploie sur Kubernetes
	@echo "$(GREEN)🚢 Déploiement sur Kubernetes...$(NC)"
	@kubectl apply -f $(MANIFESTS_DIR)/
	@echo "$(GREEN)✅ Déploiement terminé$(NC)"

undeploy: ## Supprime le déploiement
	@echo "$(YELLOW)🗑️  Suppression du déploiement...$(NC)"
	@if [ -d "$(MANIFESTS_DIR)" ]; then \
		kubectl delete -f $(MANIFESTS_DIR)/ || true; \
	fi
	@echo "$(GREEN)✅ Suppression terminée$(NC)"

status: ## Affiche le statut des pods
	@echo "$(GREEN)📊 Statut des services:$(NC)"
	@kubectl get pods,svc,pvc,ingress -n teleflix 2>/dev/null || echo "$(RED)Namespace 'teleflix' non trouvé$(NC)"

logs: ## Affiche les logs (SERVICE=nom_du_service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)❌ Spécifiez un service: make logs SERVICE=jellyfin$(NC)"; \
	else \
		echo "$(GREEN)📋 Logs de $(SERVICE):$(NC)"; \
		kubectl logs -n teleflix -l app=$(SERVICE) --tail=100 -f; \
	fi

port-forward: ## Port-forward un service (SERVICE=nom PORT=port)
	@if [ -z "$(SERVICE)" ] || [ -z "$(PORT)" ]; then \
		echo "$(RED)❌ Usage: make port-forward SERVICE=jellyfin PORT=8096$(NC)"; \
	else \
		echo "$(GREEN)🔗 Port-forward $(SERVICE) sur localhost:$(PORT)$(NC)"; \
		kubectl port-forward -n teleflix svc/$(SERVICE) $(PORT):$(PORT); \
	fi

dev-setup: ## Configure l'environnement de développement
	@echo "$(GREEN)⚙️  Configuration de l'environnement de dev...$(NC)"
	@go mod tidy
	@go mod download
	@echo "$(GREEN)✅ Environnement prêt$(NC)"

# Exemples d'utilisation rapide
examples: ## Affiche des exemples d'utilisation
	@echo "$(GREEN)📚 Exemples d'utilisation:$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Génération basique:$(NC)"
	@echo "   make generate"
	@echo ""
	@echo "$(YELLOW)2. Génération avec config personnalisée:$(NC)"
	@echo "   make generate-custom CONFIG=my-config.yaml"
	@echo ""
	@echo "$(YELLOW)3. Déploiement complet:$(NC)"
	@echo "   make deploy"
	@echo ""
	@echo "$(YELLOW)4. Port-forward Jellyfin:$(NC)"
	@echo "   make port-forward SERVICE=jellyfin PORT=8096"
	@echo ""
	@echo "$(YELLOW)5. Voir les logs de Sonarr:$(NC)"
	@echo "   make logs SERVICE=sonarr"