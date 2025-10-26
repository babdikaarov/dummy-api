import { Injectable } from '@nestjs/common';
import { db } from '../database/database';
import { locations, gates, userLocationGates } from '../database/schema';
import { eq, and } from 'drizzle-orm';

@Injectable()
export class LocationsService {
  async getAllLocations(phone?: string) {
    let allLocations;

    if (phone) {
      // Filter locations by phone number
      allLocations = await this.getLocationsByPhone(phone);
    } else {
      // Get all locations
      allLocations = await db.select().from(locations);
    }
    // Fetch gates for each location
    const locationsWithGates = await Promise.all(
      allLocations.map(async (location) => {
        const locationGates = await db
          .select({
            id: gates.id,
            title: gates.title,
            description: gates.description,
            location_id: gates.locationId,
            is_open: gates.isOpen,
            gate_is_horizontal: gates.gateIsHorizontal,
          })
          .from(gates)
          .where(eq(gates.locationId, location.id));

        return {
          id: location.id,
          title: location.title,
          address: location.address,
          logo: location.logo,
          gates: locationGates,
        };
      }),
    );

    return locationsWithGates;
  }

  async getLocationById(locationId: number) {
    return db.select().from(locations).where(eq(locations.id, locationId));
  }

  async getGatesByLocationId(locationId: number) {
    return db.select().from(gates).where(eq(gates.locationId, locationId));
  }

  async getLocationsByPhone(phone: string) {
    const result = await db
      .selectDistinct({
        id: locations.id,
        title: locations.title,
        address: locations.address,
        logo: locations.logo,
      })
      .from(userLocationGates)
      .innerJoin(locations, eq(userLocationGates.locationId, locations.id))
      .where(eq(userLocationGates.phone, phone));

    return result;
  }

  async getGatesByLocationIdAndPhone(locationId: number, phone: string) {
    return db
      .select({
        id: gates.id,
        title: gates.title,
        description: gates.description,
        location_id: gates.locationId,
        is_open: gates.isOpen,
        gate_is_horizontal: gates.gateIsHorizontal,
      })
      .from(userLocationGates)
      .innerJoin(gates, eq(userLocationGates.gateId, gates.id))
      .where(
        and(
          eq(userLocationGates.locationId, locationId),
          eq(userLocationGates.phone, phone),
        ),
      );
  }

  async openGate(gateId: number): Promise<boolean> {
    console.log(gateId, 'is attempted to be opened');
    // Simulate network delay
    await new Promise((resolve) =>
      setTimeout(resolve, 500 + Math.random() * 500),
    );
    // Return random boolean
    return Math.random() > 0.5;
  }

  async closeGate(gateId: number): Promise<boolean> {
    console.log(gateId, 'is attempted to be closed');
    // Simulate network delay
    await new Promise((resolve) =>
      setTimeout(resolve, 500 + Math.random() * 500),
    );
    // Return random boolean
    return Math.random() < 0.5;
  }

  async assignUserToLocationsAndGates(
    phone: string,
    locations: Array<{ locationId: number; gateIds: number[] }>,
  ) {
    // Delete existing assignments for this phone
    await db
      .delete(userLocationGates)
      .where(eq(userLocationGates.phone, phone));

    // Insert new assignments
    const assignments = [];
    for (const location of locations) {
      for (const gateId of location.gateIds) {
        assignments.push({
          phone,
          locationId: location.locationId,
          gateId,
        });
      }
    }

    if (assignments.length > 0) {
      await db.insert(userLocationGates).values(assignments);
    }

    return { success: true, phone };
  }
}
