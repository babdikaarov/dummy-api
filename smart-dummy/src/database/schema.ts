import {
  pgTable,
  text,
  integer,
  boolean,
  timestamp,
  primaryKey,
} from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';

// Locations table
export const locations = pgTable('locations', {
  id: integer('id').primaryKey(),
  title: text('title').notNull(),
  address: text('address').notNull(),
  logo: text('logo').notNull(),
  createdAt: timestamp('created_at').defaultNow(),
  updatedAt: timestamp('updated_at').defaultNow(),
});

// Gates table
export const gates = pgTable('gates', {
  id: integer('id').primaryKey(),
  title: text('title').notNull(),
  description: text('description').notNull(),
  locationId: integer('location_id')
    .notNull()
    .references(() => locations.id),
  isOpen: boolean('is_open').default(false),
  gateIsHorizontal: boolean('gate_is_horizontal').default(true),
  createdAt: timestamp('created_at').defaultNow(),
  updatedAt: timestamp('updated_at').defaultNow(),
});

// User-Location-Gate junction table with composite primary key (phone, location_id, gate_id)
export const userLocationGates = pgTable(
  'user_location_gates',
  {
    phone: text('phone').notNull(),
    locationId: integer('location_id')
      .notNull()
      .references(() => locations.id),
    gateId: integer('gate_id')
      .notNull()
      .references(() => gates.id),
    createdAt: timestamp('created_at').defaultNow(),
    updatedAt: timestamp('updated_at').defaultNow(),
  },
  (table) => ({
    pk: primaryKey({ columns: [table.phone, table.locationId, table.gateId] }),
  }),
);

// Relations
export const locationsRelations = relations(locations, ({ many }) => ({
  gates: many(gates),
  userLocationGates: many(userLocationGates),
}));

export const gatesRelations = relations(gates, ({ one, many }) => ({
  location: one(locations, {
    fields: [gates.locationId],
    references: [locations.id],
  }),
  userLocationGates: many(userLocationGates),
}));

export const userLocationGatesRelations = relations(
  userLocationGates,
  ({ one }) => ({
    location: one(locations, {
      fields: [userLocationGates.locationId],
      references: [locations.id],
    }),
    gate: one(gates, {
      fields: [userLocationGates.gateId],
      references: [gates.id],
    }),
  }),
);
