import "@testing-library/jest-dom";

if (!window.ResizeObserver) {
  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  window.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver;
}

if (!window.DOMMatrixReadOnly) {
  class DOMMatrixReadOnlyMock {
    a = 1;
    b = 0;
    c = 0;
    d = 1;
    e = 0;
    f = 0;
    m41 = 0;
    m42 = 0;
    constructor(_transform?: string) {}
  }

  window.DOMMatrixReadOnly = DOMMatrixReadOnlyMock as unknown as typeof DOMMatrixReadOnly;
}
