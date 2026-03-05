export class DeferredPromise<T> {
  resolve!: (a: T | PromiseLike<T>) => void;
  reject!: (reason?: any) => void; // eslint-disable-line @typescript-eslint/no-explicit-any
  promise: Promise<T>;

  constructor() {
    this.promise = new Promise<T>((resolve, reject) => {
      this.resolve = resolve;
      this.reject = reject;
    });
  }
}
