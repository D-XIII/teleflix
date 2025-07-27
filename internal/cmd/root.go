package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"teleflix/internal/config"
	"teleflix/internal/generator"

	"github.com/spf13/cobra"
)

var (
	configFile   string
	outputDir    string
	namespace    string
	storageClass string
)

var rootCmd = &cobra.Command{
	Use:   "teleflix",
	Short: "Générateur de manifests Kubernetes pour stack de streaming",
	Long: `Un générateur de manifests Kubernetes pour déployer une stack complète 
de streaming avec Jellyfin, Sonarr, Radarr, Jackett et qBittorrent.

Utilise un fichier de configuration YAML pour personnaliser tous les services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateManifests()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Fichier de configuration")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "./manifests", "Répertoire de sortie")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace Kubernetes")
	rootCmd.Flags().StringVarP(&storageClass, "storage-class", "s", "", "Classe de stockage")
}

func Execute() error {
	return rootCmd.Execute()
}

func generateManifests() error {
	// Charger la configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("erreur lors du chargement de la configuration: %w", err)
	}

	// Override avec les flags
	if namespace != "" {
		cfg.Namespace = namespace
	}
	if storageClass != "" {
		cfg.StorageClass = storageClass
	}

	// Créer le répertoire de sortie
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("erreur lors de la création du répertoire: %w", err)
	}

	// Générer les manifests
	gen := generator.New(cfg)
	manifests, err := gen.GenerateAll()
	if err != nil {
		return fmt.Errorf("erreur lors de la génération: %w", err)
	}

	// Écrire les fichiers
	for name, content := range manifests {
		filePath := filepath.Join(outputDir, name+".yaml")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("erreur lors de l'écriture de %s: %w", filePath, err)
		}
		fmt.Printf("✓ Généré: %s\n", filePath)
	}

	fmt.Printf("\n🎉 Tous les manifests ont été générés dans %s\n", outputDir)
	fmt.Println("\nPour déployer:")
	fmt.Printf("kubectl apply -f %s/\n", outputDir)

	return nil
}
