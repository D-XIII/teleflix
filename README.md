# 🎬 Teleflix - Générateur Kubernetes

Un générateur de manifests Kubernetes en Go pour déployer une stack complète de streaming avec Jellyfin, Sonarr, Radarr, Jackett et qBittorrent.

## 🚀 Fonctionnalités

- **Configuration centralisée** : Fichier YAML unique pour tous les services
- **Manifests optimisés** : Génération automatique des ressources Kubernetes
- **Stockage persistant** : Gestion des PVC pour les données et médias
- **Ingress intégré** : Configuration automatique avec TLS
- **CLI intuitive** : Commandes simples pour tous les cas d'usage

## 📋 Services inclus

| Service | Port | Description |
|---------|------|-------------|
| **Jellyfin** | 8096 | Serveur de streaming multimédia |
| **Sonarr** | 8989 | Gestionnaire de séries TV |
| **Radarr** | 7878 | Gestionnaire de films |
| **Jackett** | 9117 | Proxy pour trackers torrent |
| **qBittorrent** | 8080 | Client BitTorrent |

## 🛠️ Installation

### Prérequis
- Go 1.23+
- Kubernetes cluster
- kubectl configuré

### Compilation
```bash
# Cloner le projet
git clone <repo-url>
cd teleflix

# Installer les dépendances et compiler
make dev-setup
make build
```

## 🚀 Utilisation

### Génération des manifests
```bash
# Génération avec la config par défaut
make generate

# Génération avec une config personnalisée
make generate-custom CONFIG=my-config.yaml

# Génération avec options CLI
./bin/teleflix --namespace=media --domain=myteleflix.com
```

### Déploiement
```bash
# Déploiement automatique
make deploy

# Ou manuellement
kubectl apply -f manifests/

# Workflow complet
make all
```

### Gestion
```bash
# Voir le statut
make status

# Port-forward pour accéder aux services
make port-forward SERVICE=jellyfin PORT=8096

# Voir les logs
make logs SERVICE=sonarr

# Supprimer le déploiement
make undeploy
```

## ⚙️ Configuration

Le fichier `config.yaml` permet de personnaliser tous les aspects du déploiement :

```yaml
namespace: teleflix
domain: teleflix.local
storageClass: default

services:
  jellyfin:
    enabled: true
    image: jellyfin/jellyfin
    tag: latest
    port: 8096
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
    volumes:
      - name: media
        mountPath: /media
        readOnly: true
      - name: config
        mountPath: /config
        size: 1Gi

  sonarr:
    enabled: true
    image: linuxserver/sonarr
    tag: latest
    port: 8989
    environment:
      PUID: "1000"
      PGID: "1000"
      TZ: "Europe/Paris"
    # ... autres options
```

### Options disponibles
- **Services** : Activer/désactiver chaque service
- **Ressources** : CPU et mémoire pour chaque container
- **Stockage** : Tailles et classes de stockage
- **Ingress** : Configuration des domaines et TLS
- **Environnement** : Variables d'environnement personnalisées

## 🎯 Exemples d'usage

### Configuration minimale
```yaml
namespace: media
domain: homelab.local
services:
  jellyfin:
    enabled: true
  sonarr:
    enabled: false    # Désactiver si non nécessaire
```

### Configuration avancée
```yaml
services:
  jellyfin:
    enabled: true
    exposed: true      # Accessible publiquement via ingress
    resources:
      requests:
        cpu: 2
        memory: 4Gi
      limits:
        cpu: 4
        memory: 8Gi

  # Service activé mais non exposé publiquement
  radarr:
    enabled: true
    exposed: false     # Accessible seulement via port-forward
    # ... configuration du service
```

### Contrôle d'exposition des services
Chaque service peut être configuré indépendamment :

```yaml
services:
  jellyfin:
    enabled: true
    exposed: true      # ✅ Accessible via https://jellyfin.domain.com
    
  sonarr:
    enabled: true  
    exposed: false     # ❌ Non exposé (admin seulement)
    
  radarr:
    enabled: true
    exposed: false     # ❌ Non exposé (admin seulement)
    
  jackett:
    enabled: true
    exposed: false     # ❌ Très sensible, jamais exposé
    
  qbittorrent:
    enabled: false     # ⭕ Service complètement désactivé
```

**Accès aux services non exposés :**
```bash
# Port-forward pour accès temporaire
kubectl port-forward -n teleflix svc/sonarr 8989:8989
kubectl port-forward -n teleflix svc/radarr 7878:7878
kubectl port-forward -n teleflix svc/jackett 9117:9117
```

### Configuration avec stockage personnalisé
```yaml
storageClass: fast-ssd
storage:
  media:
    size: 500Gi
    accessModes:
      - ReadWriteMany
  downloads:
    size: 100Gi
    accessModes:
      - ReadWriteOnce
```

## 🔒 Configuration TLS/HTTPS

Teleflix intègre nativement cert-manager pour les certificats automatiques.

### Configuration Let's Encrypt (Production)
```yaml
# Dans votre config.yaml
domain: teleflix.yourdomain.com  # Votre domaine réel

ingress:
  enabled: true
  className: traefik  # ou nginx
  tls:
    enabled: true
    secretName: teleflix-tls

certManager:
  enabled: true
  issuer:
    name: teleflix-letsencrypt
    type: letsencrypt
    email: votre.email@example.com  # Requis pour Let's Encrypt
```

### Configuration Self-Signed (Tests locaux)
```yaml
# Dans votre config.yaml
domain: teleflix.local

ingress:
  enabled: true
  className: traefik
  tls:
    enabled: true
    secretName: teleflix-tls

certManager:
  enabled: true
  issuer:
    name: teleflix-selfsigned
    type: selfsigned  # Pas d'email requis
```

### Déploiement avec TLS
```bash
# Utiliser un exemple pré-configuré
make generate-custom CONFIG=examples/config-with-certmanager.yaml
# ou
make generate-custom CONFIG=examples/config-selfsigned.yaml

# Déployer
kubectl apply -f manifests/

# Vérifier les certificats
kubectl get certificate -n teleflix
kubectl get clusterissuer
```

### Vérification HTTPS
Une fois déployé, vos services seront accessibles via HTTPS :
- **Jellyfin** : `https://jellyfin.yourdomain.com`
- **Sonarr** : `https://sonarr.yourdomain.com`
- **Radarr** : `https://radarr.yourdomain.com`
- **Jackett** : `https://jackett.yourdomain.com`
- **qBittorrent** : `https://qbittorrent.yourdomain.com`

**Note** : Pour Let's Encrypt, votre domaine doit pointer vers votre cluster et être accessible depuis Internet.

## 🔧 Développement

### Ajouter un nouveau service
1. Étendre la structure `ServiceConfig` dans `config/config.go`
2. Ajouter la logique de génération dans `generator/generator.go`
3. Mettre à jour la configuration par défaut

### Tests
```bash
make test
```

### Commandes utiles
```bash
# Compilation rapide
make build

# Génération + Déploiement
make deploy

# Surveillance des logs
make logs SERVICE=jellyfin

# Accès local aux services
make port-forward SERVICE=jellyfin PORT=8096
```

## 🎮 Utilisation rapide

### 1. Démarrage rapide
```bash
git clone <repo>
cd teleflix
make dev-setup
make deploy
```

### 2. Accès Jellyfin
```bash
make port-forward SERVICE=jellyfin PORT=8096
# Ouvrir http://localhost:8096
```

### 3. Configuration des trackers
```bash
make port-forward SERVICE=jackett PORT=9117
# Configurer les trackers sur http://localhost:9117
```

### 4. Personnalisation
```bash
# Copier la config par défaut
cp config.yaml my-config.yaml

# Éditer selon vos besoins
vim my-config.yaml

# Déployer avec votre config
make generate-custom CONFIG=my-config.yaml
kubectl apply -f manifests/
```

## 🐳 Utilisation avec Docker

```bash
# Build de l'image
docker build -t teleflix:latest .

# Génération dans un container
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/manifests:/manifests \
  teleflix:latest --output /manifests
```

## 🔍 Debugging

### Script de diagnostic
```bash
./scripts/debug.sh
```

### Vérifications manuelles
```bash
# Vérifier les pods
kubectl get pods -n teleflix

# Vérifier les PVC
kubectl get pvc -n teleflix

# Logs détaillés
kubectl describe pod <pod-name> -n teleflix

# Événements
kubectl get events -n teleflix --sort-by='.lastTimestamp'
```

## 📚 Scripts inclus

### `scripts/deploy.sh`
Script de déploiement automatique avec validation :
```bash
./scripts/deploy.sh [CONFIG_FILE] [NAMESPACE]
```

### `scripts/debug.sh`
Script de diagnostic complet :
```bash
./scripts/debug.sh [NAMESPACE]
```

## 💡 Conseils et astuces

### Performance
- Augmentez les ressources CPU/RAM pour Jellyfin si vous avez beaucoup d'utilisateurs
- Utilisez une classe de stockage SSD pour les métadonnées
- Configurez l'accélération matérielle pour Jellyfin

### Sécurité
- Changez les mots de passe par défaut
- Configurez l'authentification sur l'ingress
- Utilisez des certificats TLS valides

### Stockage
- Séparez les médias des téléchargements
- Utilisez ReadWriteMany pour les médias partagés
- Planifiez l'espace disque en fonction de votre collection

## ❓ FAQ

**Q: Comment ajouter un nouveau service ?**
A: Modifiez `config.yaml` et ajoutez le service dans la section `services`, puis régénérez.

**Q: Comment changer le domaine ?**
A: Modifiez la valeur `domain` dans `config.yaml` et redéployez.

**Q: Les pods ne démarrent pas ?**
A: Vérifiez les ressources disponibles et la classe de stockage avec `./scripts/debug.sh`.

**Q: Comment sauvegarder les données ?**
A: Les données sont dans les PVC. Sauvegardez les volumes persistants de votre cluster.

## 📄 Licence

MIT License

## 🤝 Contribution

Les contributions sont les bienvenues ! N'hésitez pas à ouvrir des issues ou des pull requests.

---

**🎬 Teleflix - Votre solution de streaming Kubernetes, simple et efficace !**