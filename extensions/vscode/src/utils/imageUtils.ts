export const appendImages = (
  prevImages: string[],
  newImages: string[],
  maxImages: number
): string[] => {
  const combined = [...prevImages, ...newImages];
  return combined.slice(-maxImages);
};
