#!/bin/bash

# Teleflix - Script de déploiement rapide
# Usage: ./scripts/deploy.sh [CONFIG_FILE] [NAMESPACE]

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration par défaut
CONFIG_FILE=${1:-"config.yaml"}
NAMESPACE=${2:-"teleflix"}
MANIFESTS_DIR="./manifests"

# Fonctions utilitaires
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Banner
echo -e "${BLUE}"
echo "╔══════════════════════════════════════╗"
echo "║           🎬 TELEFLIX                ║"
echo "║    Déploiement Kubernetes Stack      ║"
echo "╚══════════════════════════════════════╝"
echo -e "${NC}"

# Vérification des prérequis
log_info "Vérification des prérequis..."

# Vérifier kubectl
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl n'est pas installé"
    exit 1
fi

# Vérifier la connexion au cluster
if ! kubectl cluster-info &> /dev/null; then
    log_error "Impossible de se connecter au cluster Kubernetes"
    exit 1
fi

log_success "kubectl configuré et cluster accessible"

# Vérifier si le binaire existe
if [ ! -f "./bin/teleflix" ]; then
    log_info "Compilation du binaire..."
    make build
fi

log_success "Binaire teleflix prêt"

# Vérifier le fichier de configuration
if [ ! -f "$CONFIG_FILE" ]; then
    log_warning "Fichier de configuration $CONFIG_FILE non trouvé, utilisation de la config par défaut"
    CONFIG_FILE="config.yaml"
fi

log_info "Configuration: $CONFIG_FILE"
log_info "Namespace: $NAMESPACE"

# Génération des manifests
log_info "Génération des manifests Kubernetes..."
./bin/teleflix --config "$CONFIG_FILE" --namespace "$NAMESPACE" --output "$MANIFESTS_DIR"
log_success "Manifests générés dans $MANIFESTS_DIR"

# Affichage des manifests qui seront déployés
log_info "Manifests à déployer:"
for file in "$MANIFESTS_DIR"/*.yaml; do
    if [ -f "$file" ]; then
        echo "  - $(basename "$file")"
    fi
done

# Confirmation avant déploiement
echo ""
read -p "Voulez-vous continuer le déploiement ? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warning "Déploiement annulé"
    exit 0
fi

# Déploiement
log_info "Déploiement en cours..."

# Créer le namespace d'abord
kubectl apply -f "$MANIFESTS_DIR/00-namespace.yaml"
log_success "Namespace créé/mis à jour"

# Déployer le storage
kubectl apply -f "$MANIFESTS_DIR/01-storage.yaml"
log_success "Stockage configuré"

# Attendre que les PVC soient liés
log_info "Attente de la liaison des volumes..."
kubectl wait --for=condition=Bound pvc --all -n "$NAMESPACE" --timeout=120s || log_warning "Certains PVC ne sont pas encore liés"

# Déployer les services
for file in "$MANIFESTS_DIR"/*.yaml; do
    filename=$(basename "$file")
    if [[ "$filename" != "00-namespace.yaml" && "$filename" != "01-storage.yaml" && "$filename" != "99-ingress.yaml" ]]; then
        log_info "Déploiement de $filename..."
        kubectl apply -f "$file"
    fi
done

# Attendre que les pods soient prêts
log_info "Attente du démarrage des pods..."
kubectl wait --for=condition=ready pod --all -n "$NAMESPACE" --timeout=300s || log_warning "Certains pods mettent plus de temps à démarrer"

# Déployer l'ingress en dernier
if [ -f "$MANIFESTS_DIR/99-ingress.yaml" ]; then
    log_info "Configuration de l'ingress..."
    kubectl apply -f "$MANIFESTS_DIR/99-ingress.yaml"
    log_success "Ingress configuré"
fi

# Affichage du statut final
log_success "Déploiement terminé !"
echo ""
log_info "Statut des services:"
kubectl get pods,svc,pvc,ingress -n "$NAMESPACE"

echo ""
log_info "Accès aux services:"

# Récupérer les informations d'accès
DOMAIN=$(grep "domain:" "$CONFIG_FILE" | awk '{print $2}' || echo "teleflix.local")

echo "  🎬 Jellyfin:    https://jellyfin.$DOMAIN"
echo "  📺 Sonarr:      https://sonarr.$DOMAIN"
echo "  🎭 Radarr:      https://radarr.$DOMAIN"
echo "  🔍 Jackett:     https://jackett.$DOMAIN"
echo "  📥 qBittorrent: https://qbittorrent.$DOMAIN"

echo ""
log_info "Pour l'accès local (port-forward):"
echo "  make port-forward SERVICE=jellyfin PORT=8096"
echo "  make port-forward SERVICE=sonarr PORT=8989"
echo "  make port-forward SERVICE=radarr PORT=7878"
echo "  make port-forward SERVICE=jackett PORT=9117"
echo "  make port-forward SERVICE=qbittorrent PORT=8080"

echo ""
log_info "Pour voir les logs:"
echo "  make logs SERVICE=jellyfin"

echo ""
log_success "🎉 Teleflix est maintenant déployé et prêt à l'emploi !"

# Nettoyage optionnel
read -p "Voulez-vous supprimer les manifests générés ? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf "$MANIFESTS_DIR"
    log_success "Manifests nettoyés"
fi