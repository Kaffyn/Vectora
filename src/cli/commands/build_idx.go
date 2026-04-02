package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/core/parser"
	"github.com/Kaffyn/Vectora/src/lib/chromem"
	"github.com/Kaffyn/Vectora/src/lib/llama"
	"github.com/spf13/cobra"
)

var (
	sourceDir  string
	outputPath string
	idxID      string
	embedURL   string
)

func init() {
	buildIdxCmd.Flags().StringVarP(&sourceDir, "source", "s", "", "Diretório com os arquivos técnicos (XML/MD/GD)")
	buildIdxCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Caminho do arquivo .idx de saída")
	buildIdxCmd.Flags().StringVarP(&idxID, "id", "i", "engine-docs", "ID único para este índice")
	buildIdxCmd.Flags().StringVarP(&embedURL, "embed-url", "e", "http://localhost:8082", "URL do sidecar de embedding")

	buildIdxCmd.MarkFlagRequired("source")
	buildIdxCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(buildIdxCmd)
}

var buildIdxCmd = &cobra.Command{
	Use:   "build-idx",
	Short: "Compila um diretório técnico em um arquivo de índice vetorial .idx",
	Run: func(cmd *cobra.Command, args []string) {
		runBuildIdx()
	},
}

func runBuildIdx() {
	ctx := context.Background()
	p := parser.NewTechnicalParser()
	embedder := llama.NewLlamaEmbedder(embedURL)
	repo := chromem.NewVectorRepo()

	fmt.Printf("🔨 Iniciando compilação do índice [%s]...\n", idxID)
	fmt.Printf("📂 Origem: %s\n", sourceDir)

	count := 0
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		doc := &domain.Document{
			ID:       filepath.Base(path),
			FilePath: path,
			Content:  string(content),
		}

		chunks, err := p.Parse(doc)
		if err != nil {
			return nil
		}

		for _, chunk := range chunks {
			fmt.Printf("  ✨ Processando chunk: %s\n", chunk.ID)

			// Gerar Embedding real via Sidecar
			emb, err := embedder.Generate(ctx, chunk.Content)
			if err != nil {
				log.Printf("❌ Erro ao gerar embedding para %s: %v", chunk.ID, err)
				continue
			}

			if err := repo.SaveToIndex(ctx, idxID, chunk, emb); err != nil {
				log.Printf("❌ Erro ao salvar chunk %s: %v", chunk.ID, err)
			}
			count++
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Fatal error during build: %v", err)
	}

	fmt.Printf("💾 Exportando %d chunks para %s...\n", count, outputPath)
	if err := repo.PersistIndex(ctx, idxID, outputPath); err != nil {
		log.Fatalf("Falha ao persistir índice: %v", err)
	}

	fmt.Println("✅ Índice .idx gerado com sucesso!")
}
