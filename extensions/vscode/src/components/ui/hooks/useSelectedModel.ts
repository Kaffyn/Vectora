/**
 * Hook para seleção de modelo
 * Placeholder para Phase 2 - Será implementado com tipos Vectora
 */

export const useSelectedModel = (apiConfiguration?: any) => {
  return {
    selectedModel: null,
    isLoading: false,
    isError: false,
    setSelectedModel: () => {},
  };
};
