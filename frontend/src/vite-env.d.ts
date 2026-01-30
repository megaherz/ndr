/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly DEV: boolean
  readonly PROD: boolean
  readonly MODE: string
  // Add more env variables here as needed
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}