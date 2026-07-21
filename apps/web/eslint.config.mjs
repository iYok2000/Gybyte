// ESLint flat config for @mono-repo/web.
// eslint-config-next v16 ships native flat-config arrays, so we spread them
// directly instead of routing through FlatCompat (which throws a circular
// "Converting circular structure to JSON" error under ESLint 9 + the Next
// plugin's self-referential config).
import nextCoreWebVitals from "eslint-config-next/core-web-vitals";
import nextTypescript from "eslint-config-next/typescript";

const eslintConfig = [
  ...nextCoreWebVitals,
  ...nextTypescript,
  {
    ignores: [".next/**", "node_modules/**"],
  },
];

export default eslintConfig;
