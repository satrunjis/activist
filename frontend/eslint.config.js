import tseslint from "typescript-eslint";
import reactHooksPlugin from "eslint-plugin-react-hooks";

export default tseslint.config(
  {
    ignores: [
      "dist/**",
      "node_modules/**",
      "coverage/**",
      "eslint.config.js",
      "postcss.config.js",
      "vite.config.ts",
      "vitest.config.ts"
    ]
  },
  ...tseslint.configs.recommended,
  {
    plugins: {
      "react-hooks": reactHooksPlugin
    },
    rules: {
      ...reactHooksPlugin.configs.recommended.rules,
      // Disable set-state-in-effect: void asyncFunc() inside useEffect is valid
      "react-hooks/set-state-in-effect": "off",
      // incompatible-library fires on react-hook-form's watch() due to React Compiler; project doesn't use Compiler
      "react-hooks/incompatible-library": "off",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" }
      ],
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/consistent-type-imports": "error",
      // no-misused-promises fires on async React event handlers; suppress for JSX attribute callbacks
      "@typescript-eslint/no-misused-promises": "off"
    }
  }
);
