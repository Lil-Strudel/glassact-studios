export interface hasMetadata {
  __hasMetadata?: true;
}

export interface hasDoubleID {
  __hasDoubleId?: true;
}

export type StandardTable = hasMetadata & hasDoubleID;

export interface Metadata {
  created_at: string;
  updated_at: string;
  version: number;
}

export interface DoubleID {
  id: number;
  uuid: string;
}

/*
 * These recursive types will check to see if type Type
 * eextends type Condition, and if it does, it will
 * return type Type & Addition. It does this
 * recursivley (duh)
 */
type Recurse<T, C, A> = {
  [K in keyof T]: T[K] extends C
    ? Recurse<T[K], C, A> & A
    : T[K] extends Object
      ? Recurse<T[K], C, A>
      : T[K];
};

type Recursive<Type, Condition, Addition> = Type extends Condition
  ? Recurse<Type, Condition, Addition> & Addition
  : Recurse<Type, Condition, Addition>;

type AddMetadata<T> = Recursive<T, hasMetadata, Metadata>;
type AddDoubleID<T> = Recursive<T, hasDoubleID, DoubleID>;

export type GET<T> = AddDoubleID<AddMetadata<T>>;
export type POST<T> = T;
export type PATCH<T> = AddDoubleID<Partial<T>>;
export type PUT<T> = AddDoubleID<T>;
