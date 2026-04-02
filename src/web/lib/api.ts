export interface SearchResult {
  content: string;
}

export interface ApiResponse<T> {
  status: number;
  results?: T[];
  error?: string;
}

export async function handleSearch(
  query: string,
): Promise<ApiResponse<SearchResult>> {
  if (!query) {
    return { status: 400, error: "Query required" };
  }

  try {
    // This will eventually call the Go Backend API
    return {
      status: 200,
      results: [{ content: "Result for: " + query }],
    };
  } catch {
    return { status: 500, error: "Internal Server Error" };
  }
}
