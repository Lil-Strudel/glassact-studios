import type {
  GetTagMetadata,
  SimplifyDeep,
  Tagged,
  UnwrapTagged,
} from "type-fest";
import { NonRecursiveType } from "type-fest/source/internal";
import type { TagContainer } from "type-fest/source/tagged";

export type { SimplifyDeep };
export type { OmitDeep, Simplify } from "type-fest";

interface Metadata {
  created_at: string;
  updated_at: string;
  version: number;
}

interface DoubleId {
  id: number;
  uuid: string;
}

export type hasMetadata<T> = Tagged<T, "Metadata", Metadata>;
export type hasDoubleId<T> = Tagged<T, "DoubleId", DoubleId>;
export type StandardTable<T> = hasMetadata<hasDoubleId<T>>;

type Tag<Token extends PropertyKey, TagMetadata> = TagContainer<{
  [K in Token]: TagMetadata;
}>;

type _addMetadata<T> =
  T extends Tag<"Metadata", unknown> ? T & GetTagMetadata<T, "Metadata"> : T;
type _addDoubleId<T> =
  T extends Tag<"DoubleId", unknown> ? GetTagMetadata<T, "DoubleId"> & T : T;

type safeUnwrap<T> = T extends Tag<PropertyKey, unknown> ? UnwrapTagged<T> : T;
// type addMetadata<T> = safeUnwrap<_addMetadata<T>>;
type addDoubleId<T> = safeUnwrap<_addDoubleId<T>>;
type addBoth<T> = safeUnwrap<_addDoubleId<_addMetadata<T>>>;

type Exclude = never | NonRecursiveType | Set<unknown> | Map<unknown, unknown>;
type deepAddDoubleId<Type> = Type extends Exclude
  ? Type
  : Type extends object
    ? {
        [TypeKey in keyof Type]: deepAddDoubleId<
          safeUnwrap<addDoubleId<Type[TypeKey]>>
        >;
      }
    : Type;

type deepAddBoth<Type> = Type extends Exclude
  ? Type
  : Type extends object
    ? {
        [TypeKey in keyof Type]: deepAddBoth<
          safeUnwrap<addBoth<Type[TypeKey]>>
        >;
      }
    : Type;

type deepSafeUnwrap<Type> = Type extends Exclude
  ? Type
  : Type extends object
    ? { [TypeKey in keyof Type]: deepSafeUnwrap<safeUnwrap<Type[TypeKey]>> }
    : Type;

export type GET<T> = SimplifyDeep<deepAddBoth<addBoth<T>>>;
export type POST<T> = SimplifyDeep<deepSafeUnwrap<safeUnwrap<T>>>;
export type PATCH<T> = SimplifyDeep<Partial<deepAddDoubleId<addDoubleId<T>>>>;
export type PUT<T> = SimplifyDeep<deepAddDoubleId<addDoubleId<T>>>;
