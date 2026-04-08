package handlers

import (
	// pb "vectora/core/api/grpc/proto"
	"vectora/core/engine"
)

type QueryHandler struct {
	// pb.UnimplementedVectoraServiceServer
	Engine *engine.Engine
}

// Mocado temporariamente ate gerarmos os arquivos PBs nativos.
/*
func (h *QueryHandler) Query(req *pb.QueryRequest, stream pb.VectoraService_QueryServer) error {
	// Usa o engine para fazer RAG e streamar tokens
	resultStream, err := h.Engine.StreamQuery(stream.Context(), req.Query, req.WorkspaceId)
	if err != nil {
		return err
	}

	for chunk := range resultStream {
		if err := stream.Send(&pb.QueryResponse{
			Token:   chunk.Token,
			Sources: chunk.Sources,
			IsFinal: chunk.IsFinal,
		}); err != nil {
			return err
		}
	}
	return nil
}
*/
