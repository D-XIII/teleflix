# 🚀 Guide de Démarrage Rapide

Ce guide vous permettra de déployer Teleflix en moins de 10 minutes !

## ⚡ Démarrage Express (2 minutes)

```bash
# 1. Cloner et compiler
git clone <repo-url>
cd teleflix
make dev-setup && make build

# 2. Déployer avec la config par défaut
make deploy

# 3. Accéder à Jellyfin
make port-forward SERVICE=jellyfin PORT=8096
# Ouvrir http://localhost:8096
```

## 🎯 Démarrage Personnalisé (5 minutes)

### 1. Choisir votre configuration

**Pour débuter (Jellyfin seulement) :**
```bash
cp examples/config-minimal.yaml my-config.yaml
```

**Pour une installation complète :**
```bash
cp config.yaml my-config.yaml
```

**Pour la production :**
```bash
cp examples/config-advanced.yaml my-config.yaml
```

### 2. Personnaliser selon vos besoins

```bash
vim my-config.yaml
```

Changez au minimum :
- `namespace`: votre nom préféré
- `domain`: votre domaine
- `storage.media.size`: selon votre espace disque

### 3. Déployer

```bash
make generate-custom CONFIG=my-config.yaml
kubectl apply -f manifests/
```

## 🔧 Configuration Rapide des Services

### Jellyfin (Obligatoire)
```bash
# Accès local
make port-forward SERVICE=jellyfin PORT=8096

# Dans le navigateur (http://localhost:8096)
1. Créer un compte administrateur
2. Ajouter les bibliothèques média :
   - Films : /media/movies
   - Séries : /media/tv
3. Configurer les métadonnées (TMDB, TVDB)
```

### Jackett (Optionnel - pour les trackers)
```bash
# Accès local
make port-forward SERVICE=jackett PORT=9117

# Dans le navigateur (http://localhost:9117)
1. Ajouter vos trackers privés/publics
2. Tester les connexions
3. Noter l'API Key pour Sonarr/Radarr
```

### Sonarr (Optionnel - pour les séries)
```bash
# Accès local
make port-forward SERVICE=sonarr PORT=8989

# Configuration de base
1. Settings > Media Management
   - Root Folder : /tv
2. Settings > Download Clients
   - Ajouter qBittorrent (http://qbittorrent:8080)
3. Settings > Indexers
   - Ajouter Jackett (http://jackett:9117 + API Key)
```

### Radarr (Optionnel - pour les films)
```bash
# Accès local
make port-forward SERVICE=radarr PORT=7878

# Configuration similaire à Sonarr
1. Root Folder : /movies
2. Même configuration download clients et indexers
```

### qBittorrent (Optionnel - pour les téléchargements)
```bash
# Accès local
make port-forward SERVICE=qbittorrent PORT=8080

# Configuration
1. Login : admin / adminadmin (changer le mot de passe !)
2. Settings > Downloads
   - Default Save Path : /downloads
3. Settings > Web UI
   - Configurer l'authentification
```

## 📁 Structure des Données Recommandée

Organisez vos médias ainsi :
```
/media/
├── movies/
│   ├── Film 1 (2020)/
│   │   └── Film 1 (2020).mkv
│   └── Film 2 (2021)/
│       └── Film 2 (2021).mkv
└── tv/
    ├── Serie 1/
    │   ├── Season 01/
    │   │   ├── S01E01.mkv
    │   │   └── S01E02.mkv
    │   └── Season 02/
    └── Serie 2/

/downloads/
├── complete/
└── incomplete/
```

## 🛠️ Dépannage Rapide

### Les pods ne démarrent pas
```bash
# Vérifier les événements
kubectl get events -n teleflix --sort-by='.lastTimestamp'

# Vérifier les ressources
kubectl describe pod <pod-name> -n teleflix

# Script de diagnostic automatique
./scripts/debug.sh
```

### Problèmes de stockage
```bash
# Vérifier les PVC
kubectl get pvc -n teleflix

# Si les PVC sont en "Pending"
kubectl describe pvc <pvc-name> -n teleflix
# → Vérifier la classe de stockage disponible
kubectl get storageclass
```

### Accès réseau
```bash
# Tester la connectivité interne
kubectl exec -it <pod-name> -n teleflix -- ping jellyfin

# Vérifier l'ingress
kubectl get ingress -n teleflix
kubectl describe ingress teleflix-ingress -n teleflix
```

## 🎛️ Commandes Utiles

```bash
# Voir tous les services
make status

# Redémarrer un service
kubectl rollout restart deployment jellyfin -n teleflix

# Mettre à jour une configuration
make generate-custom CONFIG=my-config.yaml
kubectl apply -f manifests/

# Sauvegarder la configuration
kubectl get configmap,secret -n teleflix -o yaml > backup.yaml

# Supprimer complètement
make undeploy
```

## 🎯 Prochaines Étapes

Une fois Teleflix fonctionnel :

1. **Configurer les domaines** : Pointer vos DNS vers l'ingress
2. **Ajouter du contenu** : Organiser vos médias dans la structure recommandée
3. **Optimiser** : Ajuster les ressources selon l'utilisation
4. **Sécuriser** : Configurer l'authentification et les certificats TLS
5. **Monitorer** : Surveiller les logs et performances

## 📚 Ressources Utiles

- [Configuration complète](../README.md#configuration)
- [Exemples avancés](../examples/)
- [Scripts de maintenance](../scripts/)
- [Troubleshooting](./troubleshooting.md)
