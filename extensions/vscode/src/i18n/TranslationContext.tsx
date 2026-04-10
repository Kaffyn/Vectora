import React, { createContext, useContext, ReactNode, useEffect, useCallback, useState } from "react"

let globalTranslations: Record<string, any> = {};

export const TranslationContext = createContext<{
    t: (key: string, options?: Record<string, any>) => string;
    i18n: any;
}>({
    t: (key: string) => key,
    i18n: { language: 'en', changeLanguage: async () => {} },
});

export const Trans: React.FC<{
    i18nKey: string;
    components?: Record<string, React.ReactElement>;
    values?: Record<string, any>;
}> = ({ i18nKey, components, values }) => {
    const { t } = useContext(TranslationContext);
    const text = t(i18nKey, values);

    if (!components) return <>{text}</>;

    // Basic component interpolation
    const parts = text.split(/(<[^>]+>[^<]*<\/[^>]+>)/g);
    return (
        <>
            {parts.map((part, i) => {
                const match = part.match(/<([^>]+)>(.*)<\/\1>/);
                if (match) {
                    const [_, tagName, content] = match;
                    const Component = components[tagName];
                    if (Component) {
                        return React.cloneElement(Component, { key: i }, content);
                    }
                }
                return part;
            })}
        </>
    );
};

export const TranslationProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [lang, setLang] = useState('en');

    useEffect(() => {
        const handleMessage = (event: MessageEvent) => {
            const data = event.data;
            if (data.type === 'translations') {
                globalTranslations = data.translations;
                console.log('Translations loaded in Context:', Object.keys(globalTranslations).length, 'keys');
            }
        };

        window.addEventListener('message', handleMessage);
        return () => window.removeEventListener('message', handleMessage);
    }, []);

    const translate = useCallback((key: string, options?: Record<string, any>) => {
        const cleanKey = key.includes(':') ? key.split(':')[1] : key;
        const translation = globalTranslations[cleanKey];
        if (translation && translation[lang]) {
            let text = translation[lang];
            if (options) {
                Object.keys(options).forEach(k => {
                    text = text.replace(`{{${k}}}`, options[k]);
                });
            }
            return text;
        }
        return cleanKey;
    }, [lang]);

    const i18n = {
        language: lang,
        changeLanguage: async (newLang: string) => { setLang(newLang); },
    };

    return (
        <TranslationContext.Provider value={{ t: translate, i18n }}>
            {children}
        </TranslationContext.Provider>
    );
};

export const useAppTranslation = () => useContext(TranslationContext);
export const useTranslation = () => {
    const context = useContext(TranslationContext);
    return { t: context.t, i18n: context.i18n };
};

// Global i18n object for non-React usage
export const i18n = {
    get language() { return 'en'; }, // Could be improved to track global state if needed
    t: (key: string, options?: Record<string, any>) => {
        const cleanKey = key.includes(':') ? key.split(':')[1] : key;
        const translation = globalTranslations[cleanKey];
        if (translation && translation['en']) {
            let text = translation['en'];
            if (options) {
                Object.keys(options).forEach(k => {
                    text = text.replace(`{{${k}}}`, options[k]);
                });
            }
            return text;
        }
        return cleanKey;
    },
    changeLanguage: async (lang: string) => { /* no-op globally for now */ }
};

export default TranslationProvider;
