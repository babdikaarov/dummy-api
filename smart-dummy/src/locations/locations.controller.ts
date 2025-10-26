import { Controller, Get, Put, Param, Body, Query } from '@nestjs/common';
import {
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiBody,
  ApiQuery,
} from '@nestjs/swagger';
import { LocationsService } from './locations.service';
import {
  LocationDto,
  GateDto,
  UserLocationGateAssignmentDto,
  LocationLiteDto,
} from '../dtos/location.dto';

@ApiTags('locations')
@Controller('locations')
export class LocationsController {
  constructor(private readonly locationsService: LocationsService) {}

  @Get()
  @ApiOperation({ summary: 'Get all locations' })
  @ApiQuery({
    name: 'phone',
    type: 'string',
    required: false,
    description:
      'Optional phone number to filter locations accessible to that phone',
  })
  @ApiResponse({
    status: 200,
    description: 'List of all locations',
    type: [LocationDto],
  })
  async getAllLocations(@Query('phone') phone?: string) {
    return this.locationsService.getAllLocations(phone);
  }

  @Get('by-phone/:phone/:locationId')
  @ApiOperation({
    summary: 'Get gates accessible to a phone number for a specific location',
  })
  @ApiParam({ name: 'phone', type: 'string' })
  @ApiParam({ name: 'locationId', type: 'number' })
  @ApiResponse({
    status: 200,
    description: 'List of gates related to the phone and location',
    type: [GateDto],
  })
  async getGatesByPhoneAndLocation(
    @Param('phone') phone: string,
    @Param('locationId') locationId: string,
  ) {
    return this.locationsService.getGatesByLocationIdAndPhone(
      parseInt(locationId),
      phone,
    );
  }

  @Get('by-phone/:phone')
  @ApiOperation({ summary: 'Get all locations accessible to a phone number' })
  @ApiParam({ name: 'phone', type: 'string' })
  @ApiResponse({
    status: 200,
    description: 'List of locations related to the phone',
    type: [LocationLiteDto],
  })
  async getLocationsByPhone(@Param('phone') phone: string) {
    return this.locationsService.getLocationsByPhone(phone);
  }

  @Put(':gateId/open')
  @ApiOperation({ summary: 'Open a gate' })
  @ApiParam({ name: 'gateId', type: 'number' })
  @ApiResponse({
    status: 200,
    description: 'Gate open result',
    schema: { type: 'boolean' },
  })
  async openGate(@Param('gateId') gateId: string) {
    const result = await this.locationsService.openGate(parseInt(gateId));
    return result;
  }

  @Put(':gateId/close')
  @ApiOperation({ summary: 'Close a gate' })
  @ApiParam({ name: 'gateId', type: 'number' })
  @ApiResponse({
    status: 200,
    description: 'Gate close result',
    schema: { type: 'boolean' },
  })
  async closeGate(@Param('gateId') gateId: string) {
    const result = await this.locationsService.closeGate(parseInt(gateId));
    return result;
  }

  @Put('phone')
  @ApiOperation({ summary: 'Assign user to locations and gates' })
  @ApiBody({
    type: UserLocationGateAssignmentDto,
    examples: {
      example: {
        value: {
          phone: '+77771234567',
          locations: [
            {
              locationId: 1,
              gateIds: [1, 2],
            },
            {
              locationId: 2,
              gateIds: [2],
            },
          ],
        },
      },
    },
  })
  @ApiResponse({
    status: 200,
    description: 'User assignment result',
  })
  async assignUserToLocationsAndGates(
    @Body()
    body: {
      phone: string;
      locations: Array<{ locationId: number; gateIds: number[] }>;
    },
  ) {
    return this.locationsService.assignUserToLocationsAndGates(
      body.phone,
      body.locations,
    );
  }
}
