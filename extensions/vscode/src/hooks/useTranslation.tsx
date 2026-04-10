import { useCallback, useEffect, useState } from 'react';

let globalTranslations: Record<string, any> = {};

/**
 * Custom hook to replace react-i18next.
 * Consumes translations injected via postMessage from the extension host.
 */
export const useTranslation = () => {
    const [theme, setTheme] = useState<'light' | 'dark'>('dark');
    const [lang, setLang] = useState('en');

    useEffect(() => {
        const handleMessage = (event: MessageEvent) => {
            const data = event.data;
            if (data.type === 'translations') {
                globalTranslations = data.translations;
                console.log('Translations loaded:', Object.keys(globalTranslations).length, 'keys');
            }
        };

        window.addEventListener('message', handleMessage);
        return () => window.removeEventListener('message', handleMessage);
    }, []);

    const t = useCallback((key: string, options?: any) => {
        // Handle namespaced keys like 'chat:apiRequest.title'
        const cleanKey = key.includes(':') ? key.split(':')[1] : key;
        
        // Lookup in globalTranslations
        const translation = globalTranslations[cleanKey];
        if (translation && translation[lang]) {
            let text = translation[lang];
            
            // Basic interpolation if needed
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

    return { t, i18n };
};

export const Trans = ({ i18nKey, children, values }: any) => {
    const { t } = useTranslation();
    return <span dangerouslySetInnerHTML={{ __html: t(i18nKey, values) }} /> || children;
};
