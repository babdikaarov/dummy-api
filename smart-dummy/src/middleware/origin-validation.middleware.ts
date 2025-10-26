import {
  Injectable,
  NestMiddleware,
  UnauthorizedException,
} from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';

@Injectable()
export class OriginValidationMiddleware implements NestMiddleware {
  use(req: Request, res: Response, next: NextFunction) {
    const allowedOrigin = process.env.OLOLO_MOBILE_GATE_API_ORIGIN;
    const requestOrigin = req.get('origin') || req.get('referer');

    if (!allowedOrigin) {
      // If no origin is configured, allow all requests (development mode)
      return next();
    }

    if (!requestOrigin) {
      // Allow requests without origin header (direct API calls)
      return next();
    }

    // Check if the request origin matches the allowed origin
    if (requestOrigin.startsWith(allowedOrigin)) {
      return next();
    }

    throw new UnauthorizedException(
      `Request from origin ${requestOrigin} is not allowed`,
    );
  }
}
