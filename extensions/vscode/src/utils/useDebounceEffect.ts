import { useEffect, DependencyList } from "react";

export const useDebounceEffect = (
  effect: () => void,
  delay: number,
  deps: DependencyList
) => {
  useEffect(() => {
    const handler = setTimeout(() => {
      effect();
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [...(deps || []), delay]);
};
