/**
 * Translation Schema - Defines all translation keys and their structure
 * Used for validation and type safety
 */

export const translationSchema = {
  // Common UI elements
  common: {
    cancel: "Cancel",
    confirm: "Confirm",
    ok: "OK",
    close: "Close",
    save: "Save",
    delete: "Delete",
    edit: "Edit",
    loading: "Loading...",
    error: "Error",
    warning: "Warning",
    success: "Success",
    info: "Information",
    yes: "Yes",
    no: "No",
    back: "Back",
    next: "Next",
    previous: "Previous",
    search: "Search",
    settings: "Settings",
    about: "About",
    help: "Help",
    required: "Required",
    optional: "Optional",
    apply: "Apply",
    reset: "Reset",
    clear: "Clear",
    copy: "Copy",
    paste: "Paste",
    version: "Version",
    language: "Language",
    theme: "Theme",
    accessibility: "Accessibility",
  },

  // Chat messages
  chat: {
    title: "Vectora Chat",
    placeholder: "Type your message here...",
    send: "Send",
    sendMessage: "Send message (Ctrl+Enter)",
    clear: "Clear history",
    clearConfirm: "Are you sure you want to clear the history?",
    noMessages: "No messages yet. Start a conversation!",
    thinking: "Thinking...",
    userLabel: "You",
    assistantLabel: "Assistant",
    errorSending: "Error sending message",
    connectionError: "Connection error with server",
    messageNotSent: "Message not sent",
    resend: "Resend",
    copy: "Copy message",
    edit: "Edit message",
    delete: "Delete message",
    reactions: "React",
    streaming: "Streaming response...",
    aborted: "Message aborted",
    tokens: "Tokens used: {{count}}",
    tokensPlural: "Tokens used: {{count}}",
  },

  // Settings
  settings: {
    title: "Settings",
    language: "Language",
    theme: "Theme",
    darkMode: "Dark Mode",
    lightMode: "Light Mode",
    highContrast: "High Contrast",
    autoDetect: "Auto-detect",
    fontSize: "Font Size",
    fontSizeSmall: "Small",
    fontSizeMedium: "Medium",
    fontSizeLarge: "Large",
    accessibility: "Accessibility",
    a11y: "Accessibility Settings",
    screenReader: "Screen reader support",
    focusIndicator: "Show focus indicators",
    reducedMotion: "Respect reduced motion preference",
    contrast: "Increased contrast",
    highContrastMode: "High contrast mode",
    codeEditor: "Code Editor",
    lineNumbers: "Line numbers",
    syntaxHighlight: "Syntax highlighting",
    autoComplete: "Auto-complete",
    workspace: "Workspace",
    workspaceSettings: "Workspace Settings",
    defaultWorkspace: "Default workspace",
    maxHistorySize: "Maximum history size",
    autoSave: "Auto-save messages",
    cacheSize: "Cache size (MB)",
  },

  // RAG (Retrieval Augmented Generation)
  rag: {
    title: "RAG Settings",
    enabled: "Enable RAG",
    files: "Files",
    documents: "Documents",
    relevance: "Relevance",
    similarity: "Similarity score",
    cosineSimilarity: "Cosine similarity",
    distance: "Distance",
    relevantFiles: "Relevant files",
    noRelevantFiles: "No relevant files found",
    loadingFiles: "Loading files...",
    filePreview: "File preview",
    showMore: "Show more lines",
    hideMore: "Hide lines",
    lineCount: "Lines: {{count}}",
    searchPlaceholder: "Search files...",
    context: "Context",
    contextSize: "Context window size",
    contextTokens: "Context tokens: {{count}}",
    indexing: "Indexing files...",
    indexed: "Files indexed: {{count}}",
  },

  // Errors
  errors: {
    generic: "An error occurred",
    networkError: "Network error",
    timeout: "Request timeout",
    fileNotFound: "File not found",
    permissionDenied: "Permission denied",
    invalidInput: "Invalid input",
    sessionExpired: "Session expired",
    reconnecting: "Reconnecting...",
    offline: "You are offline",
    retry: "Retry",
    dismiss: "Dismiss",
    details: "Error details",
  },

  // Accessibility
  a11y: {
    skipToContent: "Skip to main content",
    openMenu: "Open menu",
    closeMenu: "Close menu",
    toggleTheme: "Toggle theme",
    toggleLanguage: "Toggle language",
    focusTrap: "Focus trapped in dialog",
    announceLoading: "Loading content...",
    announceSuccess: "Operation successful",
    announceError: "An error occurred",
    announceWarning: "Warning",
    announceInfo: "Information",
    announceLiveRegion: "Live region",
    mainContent: "Main content",
    navigation: "Navigation",
    sidebar: "Sidebar",
    footer: "Footer",
  },

  // Keyboard shortcuts
  shortcuts: {
    title: "Keyboard Shortcuts",
    send: "Send message",
    newChat: "New chat",
    clearChat: "Clear chat",
    settings: "Open settings",
    search: "Search files",
    focusInput: "Focus input",
    selectAll: "Select all",
    copy: "Copy",
    paste: "Paste",
    undo: "Undo",
    redo: "Redo",
  },

  // Status messages
  status: {
    connecting: "Connecting...",
    connected: "Connected",
    disconnected: "Disconnected",
    reconnecting: "Reconnecting...",
    error: "Error",
    saving: "Saving...",
    saved: "Saved",
    syncing: "Syncing...",
    synced: "Synced",
  },

  // Validation messages
  validation: {
    required: "This field is required",
    invalidEmail: "Invalid email address",
    invalidUrl: "Invalid URL",
    tooShort: "Too short (minimum {{min}} characters)",
    tooLong: "Too long (maximum {{max}} characters)",
    passwordWeak: "Password is weak",
    passwordStrong: "Password is strong",
    fileNotAllowed: "File type not allowed",
    fileTooLarge: "File is too large",
  },
} as const;

/**
 * Language metadata
 */
export const languageMetadata = {
  "en-US": {
    name: "English (US)",
    nativeName: "English",
    direction: "ltr",
    region: "US",
    localeCode: "en-US",
    dateFormat: "MM/DD/YYYY",
    timeFormat: "h:mm A",
    numberFormat: "1,234.56",
  },
  "pt-BR": {
    name: "Português (Brasil)",
    nativeName: "Português",
    direction: "ltr",
    region: "BR",
    localeCode: "pt-BR",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "HH:mm",
    numberFormat: "1.234,56",
  },
  "es-ES": {
    name: "Español (España)",
    nativeName: "Español",
    direction: "ltr",
    region: "ES",
    localeCode: "es-ES",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "HH:mm",
    numberFormat: "1.234,56",
  },
} as const;

/**
 * Type for valid language codes
 */
export type LanguageCode = keyof typeof languageMetadata;

/**
 * Type for translation paths
 */
export type TranslationPath = keyof typeof translationSchema;
