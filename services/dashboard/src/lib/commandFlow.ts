import type { Component } from 'vue';
import type { DocumentNode } from 'graphql';

export type FlowStep = {
  id: string;
  placeholder: string;
  initial?: string;
  validate?: (value: string) => boolean;
  invalidHint?: string;
  required?: boolean;
};

export type CommandFlowConfig = {
  title: string;
  iconSrc?: string;
  icon?: Component;
  submitLabel?: string;
  steps: FlowStep[];
  mutation: DocumentNode;
  variables: (values: Record<string, string>) => Record<string, unknown>;
};
