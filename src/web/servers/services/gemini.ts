import { ChatGoogleGenerativeAI } from "@langchain/google-genai";
import { SystemMessage, HumanMessage } from "@langchain/core/messages";

/**
 * Gemini Service (Cloud Premium Mode).
 * 
 * Uses LangChain to interact with Gemini API. 
 * Expected to receive API key via env or parameter.
 */
export async function handleGeminiChat(message: string, apiKey?: string) {
  const finalApiKey = apiKey || process.env.GEMINI_API_KEY;

  if (!finalApiKey) {
    return {
      status: 400,
      body: { error: "GEMINI_API_KEY is not configured or provided by the client." }
    };
  }

  try {
    const model = new ChatGoogleGenerativeAI({
      modelName: "gemini-1.5-pro", 
      apiKey: finalApiKey,
      maxOutputTokens: 2048,
    });

    const sysMsg = new SystemMessage(
      "Você é o Vectora no Modo Cloud Premium. Atuando com motores do Google Gemini para alta performance e multimodalidade. " +
      "Responda à pergunta do desenvolvedor de jogos ajudando-o com sua engine/framework."
    );
    const userMsg = new HumanMessage(message);

    const result = await model.invoke([sysMsg, userMsg]);
    
    // Convert to JSON matching the Qwen response structure
    return {
      status: 200,
      body: {
        reply: result.content,
        sources: [
          {
            title: "Modo Cloud Premium",
            content: "Nenhuma fonte local carregada. Resposta gerada via Inteligência em Nuvem (Gemini Base Model).",
            path: "cloud://gemini-pro"
          }
        ]
      }
    };
  } catch (err: any) {
    return {
      status: 500,
      body: { error: "Gemini APIs error: " + err.message }
    };
  }
}
