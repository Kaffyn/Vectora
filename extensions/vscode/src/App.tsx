import { ExtensionStateContextProvider } from "./context/ExtensionStateContext";
import { TranslationProvider } from "./i18n/TranslationContext";
import ChatView from "./components/chat/ChatView";

function App() {
  return (
    <ExtensionStateContextProvider>
      <TranslationProvider>
        <div className="flex flex-col h-screen bg-vscode-bg text-vscode-fg overflow-hidden">
          <ChatView isHidden={false} showAnnouncement={false} hideAnnouncement={() => {}} />
        </div>
      </TranslationProvider>
    </ExtensionStateContextProvider>
  );
}

export default App;
