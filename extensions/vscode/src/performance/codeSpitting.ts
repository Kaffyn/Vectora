/**
 * Code Splitting & Lazy Loading Module
 * Phase 15: Performance Final Polish
 *
 * Provides:
 * - Dynamic imports and code splitting
 * - Lazy loading strategies
 * - Bundle analysis
 * - Prefetching and preloading
 */

/**
 * Module metadata
 */
export interface ModuleMetadata {
  name: string;
  size: number;
  loadTime: number;
  cached: boolean;
  exports: string[];
}

/**
 * Split strategy
 */
export type SplitStrategy = 'route' | 'component' | 'utility' | 'vendor';

/**
 * Code splitter
 */
export class CodeSplitter {
  private loadedModules: Map<string, any> = new Map();
  private loadingModules: Map<string, Promise<any>> = new Map();
  private moduleMetadata: Map<string, ModuleMetadata> = new Map();
  private eventHandlers: Map<string, Function[]> = new Map();

  /**
   * Dynamically import module
   */
  async importModule(modulePath: string): Promise<any> {
    // Return cached module
    if (this.loadedModules.has(modulePath)) {
      this.emit('module-loaded-cached', { modulePath });
      return this.loadedModules.get(modulePath);
    }

    // Return pending import
    if (this.loadingModules.has(modulePath)) {
      return this.loadingModules.get(modulePath);
    }

    // Start new import
    const startTime = Date.now();
    const importPromise = import(modulePath)
      .then(module => {
        const loadTime = Date.now() - startTime;

        // Cache module
        this.loadedModules.set(modulePath, module);
        this.loadingModules.delete(modulePath);

        // Record metadata
        this.recordMetadata(modulePath, module, loadTime);

        this.emit('module-loaded', { modulePath, loadTime });

        return module;
      })
      .catch(error => {
        this.loadingModules.delete(modulePath);
        this.emit('module-load-error', { modulePath, error });
        throw error;
      });

    this.loadingModules.set(modulePath, importPromise);

    return importPromise;
  }

  /**
   * Prefetch module (background loading)
   */
  prefetchModule(modulePath: string): void {
    // Don't prefetch if already loaded or loading
    if (this.loadedModules.has(modulePath) || this.loadingModules.has(modulePath)) {
      return;
    }

    // Prefetch in background
    this.importModule(modulePath).catch(() => {
      // Silently fail prefetch
    });

    this.emit('module-prefetch-started', { modulePath });
  }

  /**
   * Preload module (high priority)
   */
  async preloadModule(modulePath: string): Promise<void> {
    await this.importModule(modulePath);
    this.emit('module-preloaded', { modulePath });
  }

  /**
   * Batch preload modules
   */
  async preloadModules(modulePaths: string[]): Promise<void> {
    await Promise.all(modulePaths.map(path => this.importModule(path)));
    this.emit('modules-preloaded', { count: modulePaths.length });
  }

  /**
   * Get module metadata
   */
  getModuleMetadata(modulePath: string): ModuleMetadata | null {
    return this.moduleMetadata.get(modulePath) || null;
  }

  /**
   * Get all module metadata
   */
  getAllModuleMetadata(): ModuleMetadata[] {
    return Array.from(this.moduleMetadata.values());
  }

  /**
   * Clear module from cache
   */
  clearModule(modulePath: string): boolean {
    const removed = this.loadedModules.delete(modulePath);
    this.moduleMetadata.delete(modulePath);
    return removed;
  }

  /**
   * Clear all modules
   */
  clearAllModules(): void {
    this.loadedModules.clear();
    this.moduleMetadata.clear();
  }

  /**
   * Get bundle statistics
   */
  getBundleStatistics() {
    const modules = Array.from(this.moduleMetadata.values());

    return {
      totalModules: modules.length,
      totalSize: modules.reduce((sum, m) => sum + m.size, 0),
      averageLoadTime: modules.length > 0
        ? modules.reduce((sum, m) => sum + m.loadTime, 0) / modules.length
        : 0,
      cachedModules: modules.filter(m => m.cached).length,
      modules: modules.map(m => ({
        name: m.name,
        size: m.size,
        loadTime: m.loadTime
      }))
    };
  }

  /**
   * Record module metadata
   */
  private recordMetadata(modulePath: string, module: any, loadTime: number): void {
    const exports = Object.keys(module);

    const metadata: ModuleMetadata = {
      name: modulePath.split('/').pop() || modulePath,
      size: JSON.stringify(module).length,
      loadTime,
      cached: true,
      exports
    };

    this.moduleMetadata.set(modulePath, metadata);
  }

  /**
   * Register event handler
   */
  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  /**
   * Unregister event handler
   */
  off(event: string, handler: Function): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Emit event
   */
  private emit(event: string, data: any): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      for (const handler of handlers) {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in event handler for ${event}:`, error);
        }
      }
    }
  }
}

/**
 * Lazy component loader for React
 */
export class LazyComponentLoader {
  private components: Map<string, any> = new Map();
  private loadingComponents: Map<string, Promise<any>> = new Map();

  /**
   * Load component lazily
   */
  async loadComponent(componentPath: string): Promise<React.ComponentType<any>> {
    if (this.components.has(componentPath)) {
      return this.components.get(componentPath);
    }

    if (this.loadingComponents.has(componentPath)) {
      return this.loadingComponents.get(componentPath);
    }

    const promise = import(componentPath)
      .then(module => {
        const component = module.default || module;
        this.components.set(componentPath, component);
        this.loadingComponents.delete(componentPath);
        return component;
      })
      .catch(error => {
        this.loadingComponents.delete(componentPath);
        throw error;
      });

    this.loadingComponents.set(componentPath, promise);

    return promise;
  }

  /**
   * Create lazy component
   */
  createLazyComponent(componentPath: string, fallback?: React.ComponentType<any>) {
    const React = require('react');
    const Suspense = React.Suspense;

    return React.lazy(() =>
      this.loadComponent(componentPath).then(component => ({
        default: component
      }))
    );
  }

  /**
   * Prefetch component
   */
  prefetchComponent(componentPath: string): void {
    this.loadComponent(componentPath).catch(() => {
      // Silently fail prefetch
    });
  }

  /**
   * Clear component
   */
  clearComponent(componentPath: string): boolean {
    return this.components.delete(componentPath);
  }

  /**
   * Clear all components
   */
  clearAllComponents(): void {
    this.components.clear();
  }
}

/**
 * Route-based code splitting
 */
export class RouteCodeSplitter {
  private routeModules: Map<string, () => Promise<any>> = new Map();
  private preloadedRoutes: Set<string> = new Set();

  /**
   * Register route module
   */
  registerRoute(routePath: string, moduleLoader: () => Promise<any>): void {
    this.routeModules.set(routePath, moduleLoader);
  }

  /**
   * Get route module
   */
  async getRouteModule(routePath: string): Promise<any> {
    const loader = this.routeModules.get(routePath);

    if (!loader) {
      throw new Error(`No module registered for route: ${routePath}`);
    }

    return loader();
  }

  /**
   * Preload route
   */
  async preloadRoute(routePath: string): Promise<void> {
    if (this.preloadedRoutes.has(routePath)) {
      return;
    }

    await this.getRouteModule(routePath);
    this.preloadedRoutes.add(routePath);
  }

  /**
   * Preload adjacent routes
   */
  async preloadAdjacentRoutes(currentRoute: string): Promise<void> {
    // This would be implemented based on your routing structure
    // For example, preload next/previous routes in navigation flow
  }

  /**
   * Get registered routes
   */
  getRegisteredRoutes(): string[] {
    return Array.from(this.routeModules.keys());
  }

  /**
   * Get preloaded routes
   */
  getPreloadedRoutes(): string[] {
    return Array.from(this.preloadedRoutes);
  }
}

export default CodeSplitter;
