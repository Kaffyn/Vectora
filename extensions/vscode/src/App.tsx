import { ExtensionStateContextProvider } from './context/ExtensionStateContext';
import ChatView from './components/chat/ChatView';

function App() {
    return (
        <ExtensionStateContextProvider>
            <div className="flex flex-col h-screen bg-vscode-bg text-vscode-fg overflow-hidden">
                <ChatView 
                    isHidden={false} 
                    showAnnouncement={false} 
                    hideAnnouncement={() => {}} 
                />
            </div>
        </ExtensionStateContextProvider>
    );
}

export default App;
