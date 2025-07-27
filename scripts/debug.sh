#!/bin/bash

# Teleflix - Script de diagnostic et debug
# Usage: ./scripts/debug.sh [NAMESPACE]

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE=${1:-"teleflix"}

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

log_section() {
    echo -e "${CYAN}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

# Banner
echo -e "${BLUE}"
echo "╔══════════════════════════════════════╗"
echo "║         🔧 TELEFLIX DEBUG            ║"
echo "║    Diagnostic Kubernetes Stack       ║"
echo "╚══════════════════════════════════════╝"
echo -e "${NC}"

# Vérification de la connectivité cluster
log_section "🔍 VÉRIFICATION CLUSTER"
if kubectl cluster-info &> /dev/null; then
    log_success "Connexion au cluster OK"
    kubectl cluster-info
else
    log_error "Impossible de se connecter au cluster Kubernetes"
    exit 1
fi

echo ""

# Vérification du namespace
log_section "📁 VÉRIFICATION NAMESPACE"
if kubectl get namespace "$NAMESPACE" &> /dev/null; then
    log_success "Namespace '$NAMESPACE' existe"
else
    log_error "Namespace '$NAMESPACE' n'existe pas"
    echo "Créez-le avec: kubectl create namespace $NAMESPACE"
    exit 1
fi

echo ""

# Statut des ressources
log_section "📊 STATUT DES RESSOURCES"
log_info "Pods:"
kubectl get pods -n "$NAMESPACE" -o wide 2>/dev/null || log_warning "Aucun pod trouvé"

echo ""
log_info "Services:"
kubectl get svc -n "$NAMESPACE" 2>/dev/null || log_warning "Aucun service trouvé"

echo ""
log_info "PersistentVolumeClaims:"
kubectl get pvc -n "$NAMESPACE" 2>/dev/null || log_warning "Aucun PVC trouvé"

echo ""
log_info "Ingress:"
kubectl get ingress -n "$NAMESPACE" 2>/dev/null || log_warning "Aucun ingress trouvé"

echo ""

# Vérification détaillée des pods
log_section "🔍 DIAGNOSTIC DÉTAILLÉ DES PODS"
PODS=$(kubectl get pods -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)

if [ -z "$PODS" ]; then
    log_warning "Aucun pod trouvé dans le namespace $NAMESPACE"
else
    for pod in $PODS; do
        echo ""
        log_info "Analyse du pod: $pod"
        
        # Statut du pod
        STATUS=$(kubectl get pod "$pod" -n "$NAMESPACE" -o jsonpath='{.status.phase}')
        case $STATUS in
            "Running")
                log_success "Statut: $STATUS"
                ;;
            "Pending")
                log_warning "Statut: $STATUS"
                ;;
            "Failed"|"Error")
                log_error "Statut: $STATUS"
                ;;
            *)
                echo "  Statut: $STATUS"
                ;;
        esac
        
        # Conditions du pod
        kubectl get pod "$pod" -n "$NAMESPACE" -o jsonpath='{range .status.conditions[*]}{.type}{": "}{.status}{"\n"}{end}' | while read condition; do
            if [[ $condition == *"True" ]]; then
                echo -e "  ${GREEN}✓${NC} $condition"
            elif [[ $condition == *"False" ]]; then
                echo -e "  ${RED}✗${NC} $condition"
            else
                echo "  $condition"
            fi
        done
        
        # Vérifier les événements récents pour ce pod
        echo "  Événements récents:"
        kubectl get events -n "$NAMESPACE" --field-selector involvedObject.name="$pod" --sort-by='.lastTimestamp' | tail -3 | while read event; do
            echo "    $event"
        done
    done
fi

echo ""

# Vérification des PVC
log_section "💾 DIAGNOSTIC STOCKAGE"
PVCS=$(kubectl get pvc -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)

if [ -z "$PVCS" ]; then
    log_warning "Aucun PVC trouvé"
else
    for pvc in $PVCS; do
        echo ""
        log_info "PVC: $pvc"
        STATUS=$(kubectl get pvc "$pvc" -n "$NAMESPACE" -o jsonpath='{.status.phase}')
        case $STATUS in
            "Bound")
                log_success "Statut: $STATUS"
                ;;
            "Pending")
                log_warning "Statut: $STATUS - Vérifiez la classe de stockage"
                ;;
            *)
                log_error "Statut: $STATUS"
                ;;
        esac
        
        CAPACITY=$(kubectl get pvc "$pvc" -n "$NAMESPACE" -o jsonpath='{.status.capacity.storage}')
        STORAGECLASS=$(kubectl get pvc "$pvc" -n "$NAMESPACE" -o jsonpath='{.spec.storageClassName}')
        echo "  Capacité: $CAPACITY"
        echo "  Classe de stockage: $STORAGECLASS"
    done
fi

echo ""

# Vérification de l'ingress
log_section "🌐 DIAGNOSTIC INGRESS"
INGRESSES=$(kubectl get ingress -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)

if [ -z "$INGRESSES" ]; then
    log_warning "Aucun ingress trouvé"
else
    for ingress in $INGRESSES; do
        echo ""
        log_info "Ingress: $ingress"
        
        # Récupérer les hosts
        HOSTS=$(kubectl get ingress "$ingress" -n "$NAMESPACE" -o jsonpath='{.spec.rules[*].host}')
        echo "  Hosts configurés: $HOSTS"
        
        # Vérifier la classe d'ingress
        INGRESS_CLASS=$(kubectl get ingress "$ingress" -n "$NAMESPACE" -o jsonpath='{.spec.ingressClassName}')
        echo "  Classe d'ingress: $INGRESS_CLASS"
        
        # Vérifier si l'ingress controller existe
        if kubectl get ingressclass "$INGRESS_CLASS" &> /dev/null; then
            log_success "Classe d'ingress '$INGRESS_CLASS' existe"
        else
            log_error "Classe d'ingress '$INGRESS_CLASS' n'existe pas"
            echo "    Classes disponibles:"
            kubectl get ingressclass
        fi
    done
fi

echo ""

# Logs récents des services principaux
log_section "📋 LOGS RÉCENTS"
SERVICES=("jellyfin" "sonarr" "radarr" "jackett" "qbittorrent")

for service in "${SERVICES[@]}"; do
    echo ""
    log_info "Logs récents de $service:"
    
    POD=$(kubectl get pods -n "$NAMESPACE" -l app="$service" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -n "$POD" ]; then
        kubectl logs "$POD" -n "$NAMESPACE" --tail=5 2>/dev/null | while read line; do
            echo "    $line"
        done
    else
        log_warning "Aucun pod trouvé pour $service"
    fi
done

echo ""

# Conseils de dépannage
log_section "💡 CONSEILS DE DÉPANNAGE"
echo ""
echo "🔧 Commandes utiles pour le debug:"
echo ""
echo "  # Voir les logs d'un service spécifique"
echo "  kubectl logs -l app=jellyfin -n $NAMESPACE --tail=100 -f"
echo ""
echo "  # Décrire un pod pour plus d'infos"
echo "  kubectl describe pod <nom-du-pod> -n $NAMESPACE"
echo ""
echo "  # Accéder à un pod pour debug"
echo "  kubectl exec -it <nom-du-pod> -n $NAMESPACE -- /bin/bash"
echo ""
echo "  # Vérifier les événements du namespace"
echo "  kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp'"
echo ""
echo "  # Port-forward pour tester l'accès"
echo "  kubectl port-forward svc/jellyfin -n $NAMESPACE 8096:8096"
echo ""

# Problèmes courants
echo "🐛 Problèmes courants:"
echo ""
echo "  • Pods en Pending: Vérifiez les ressources et le stockage"
echo "  • ImagePullBackOff: Vérifiez la connectivité internet et les noms d'images"
echo "  • CrashLoopBackOff: Consultez les logs du pod"
echo "  • PVC en Pending: Vérifiez la classe de stockage par défaut"
echo "  • Ingress non accessible: Vérifiez l'ingress controller et le DNS"
echo ""

log_success "Diagnostic terminé !"

# Proposer des actions
echo ""
read -p "Voulez-vous voir les logs détaillés d'un service ? (service/N): " -r SERVICE
if [[ -n "$SERVICE" && "$SERVICE" != "N" && "$SERVICE" != "n" ]]; then
    POD=$(kubectl get pods -n "$NAMESPACE" -l app="$SERVICE" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -n "$POD" ]; then
        log_info "Logs détaillés de $SERVICE:"
        kubectl logs "$POD" -n "$NAMESPACE" --tail=50
    else
        log_error "Service '$SERVICE' non trouvé"
    fi
fi