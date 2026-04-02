import { z } from 'zod';

// Validador de Chat de Acordo com o Modo de IA Selecionado
export const chatRequestSchema = z.object({
  message: z.string().min(1, "Message cannot be empty"),
  conversationId: z.string().optional(),
  provider: z.enum(['qwen', 'gemini']).default('qwen'),
  model: z.string().optional(),
  // Opcional para o Gemini Mode (caso o usuário envie ou peguemos via hook web)
  api_key: z.string().optional(),
});
