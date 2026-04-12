/**
 * ChatView - Main Chat Interface for Vectora VS Code Extension
 *
 * Simplified component that connects to Vectora Core via ACP protocol
 * and provides a clean chat interface with message history and real-time streaming.
 */

import React, { forwardRef, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Virtuoso, type VirtuosoHandle } from "react-virtuoso";
import type { VectoraMessage } from "@/types/vectora";
import { useVectoraState, useSession } from "@/context";
import ChatRow from "./ChatRow";
import { ChatTextArea } from "./ChatTextArea";
import FileChangesPanel from "./FileChangesPanel";

export interface ChatViewProps {
  isHidden: boolean;
}

export interface ChatViewRef {
  acceptInput: () => void;
}

// Maximum number of images per message (Anthropic limit)
export const MAX_IMAGES_PER_MESSAGE = 20;

/**
 * Simplified ChatView component for Vectora
 */
const ChatViewComponent: React.ForwardRefRenderFunction<ChatViewRef, ChatViewProps> = ({ isHidden }) => {
  const state = useVectoraState();
  const session = useSession();

  const [inputValue, setInputValue] = useState("");
  const [selectedImages, setSelectedImages] = useState<string[]>([]);
  const textAreaRef = useRef<HTMLTextAreaElement>(null);
  const virtuosoRef = useRef<VirtuosoHandle>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  // Memoize messages to avoid unnecessary re-renders
  const messages = useMemo(() => session.messages || [], [session.messages]);

  /**
   * Handle sending a message to the AI
   */
  const handleSendMessage = useCallback(
    async (content: string) => {
      if (!content.trim() || !session.sessionId) {
        return;
      }

      try {
        setInputValue("");
        setSelectedImages([]);
        await session.sendMessage(content);
      } catch (error) {
        console.error("Failed to send message:", error);
      }
    },
    [session]
  );

  /**
   * Render individual message row
   */
  const itemContent = useCallback(
    (_index: number, message: VectoraMessage) => {
      return (
        <ChatRow
          key={message.id || _index}
          message={message}
          isLast={_index === messages.length - 1}
          isExpanded={false}
          isStreaming={state.isProcessing}
          onToggleExpand={() => {}}
          onHeightChange={() => {}}
        />
      );
    },
    [messages.length, state.isProcessing]
  );

  /**
   * Scroll to bottom when new messages arrive
   */
  useEffect(() => {
    if (virtuosoRef.current) {
      virtuosoRef.current.scrollToIndex(messages.length - 1);
    }
  }, [messages.length]);

  /**
   * Show welcome screen if no session active
   */
  if (!session.sessionId || isHidden) {
    return (
      <div className="flex flex-col h-full justify-center items-center gap-4 p-6 text-center">
        <h1 className="text-2xl font-bold">Vectora AI</h1>
        <p className="text-sm opacity-70">
          {!session.sessionId
            ? "Criar ou selecionar uma sessão para começar"
            : "Chat oculto"}
        </p>
        {session.sessionId && (
          <div className="text-xs opacity-50">
            Sessão: {session.sessionId.substring(0, 8)}...
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full w-full">
      {/* Message list */}
      <div className="grow flex min-h-0" ref={scrollContainerRef}>
        <Virtuoso
          ref={virtuosoRef}
          className="scrollable grow overflow-y-auto"
          data={messages}
          itemContent={itemContent}
          increaseViewportBy={{ top: 300, bottom: 300 }}
          atBottomThreshold={10}
        />
      </div>

      {/* File changes panel (if applicable) */}
      <FileChangesPanel clineMessages={messages} />

      {/* Status bar */}
      {state.isProcessing && (
        <div className="px-4 py-2 text-xs opacity-60 border-t border-vscode-widget-border">
          <span className="animate-pulse">Processando...</span>
        </div>
      )}

      {state.error && (
        <div className="px-4 py-2 text-xs text-red-400 border-t border-vscode-widget-border bg-red-400 bg-opacity-10">
          {state.error}
        </div>
      )}

      {/* Token usage (if available) */}
      {session.messages && session.messages.length > 0 && (
        <div className="px-4 py-1 text-xs opacity-50 border-t border-vscode-widget-border">
          {`${messages.length} mensagens`}
        </div>
      )}

      {/* Input area */}
      <ChatTextArea
        ref={textAreaRef}
        inputValue={inputValue}
        setInputValue={setInputValue}
        sendingDisabled={session.isLoading || state.isProcessing || !session.sessionId}
        onSend={() => handleSendMessage(inputValue)}
        selectedImages={selectedImages}
        setSelectedImages={setSelectedImages}
        onHeightChange={() => {
          // Auto-scroll to bottom when input height changes
          if (virtuosoRef.current) {
            virtuosoRef.current.scrollToIndex(messages.length - 1);
          }
        }}
      />
    </div>
  );
};

const ChatView = forwardRef(ChatViewComponent);

export default ChatView;
