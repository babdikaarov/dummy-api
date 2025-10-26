import { ApiProperty } from '@nestjs/swagger';

export class GateDto {
  @ApiProperty()
  id: number;

  @ApiProperty()
  title: string;

  @ApiProperty()
  description: string;

  @ApiProperty()
  location_id: number;

  @ApiProperty()
  is_open: boolean;

  @ApiProperty()
  gate_is_horizontal: boolean;
}

export class LocationDto {
  @ApiProperty()
  id: number;

  @ApiProperty()
  title: string;

  @ApiProperty()
  address: string;

  @ApiProperty()
  logo: string;

  @ApiProperty({ type: [GateDto], required: false })
  gates?: GateDto[];
}
export class LocationLiteDto {
  @ApiProperty()
  id: number;

  @ApiProperty()
  title: string;

  @ApiProperty()
  address: string;

  @ApiProperty()
  logo: string;
}

export class LocationAssignmentDto {
  @ApiProperty()
  locationId: number;

  @ApiProperty({ type: [Number] })
  gateIds: number[];
}

export class UserLocationGateAssignmentDto {
  @ApiProperty()
  phone: string;

  @ApiProperty({ type: [LocationAssignmentDto] })
  locations: LocationAssignmentDto[];
}
