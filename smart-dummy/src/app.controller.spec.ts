import { Test, TestingModule } from '@nestjs/testing';
import { AppController } from './app.controller';
import { AppService } from './app.service';

describe('AppController', () => {
  let appController: AppController;

  beforeEach(async () => {
    const app: TestingModule = await Test.createTestingModule({
      controllers: [AppController],
      providers: [AppService],
    }).compile();

    appController = app.get<AppController>(AppController);
  });

  describe('getHealthCheck', () => {
    it('should return health check object with status ok', () => {
      const result = appController.getHealthCheck();
      expect(result).toBeDefined();
      expect(result.status).toBe('ok');
      expect(result.timestamp).toBeDefined();
      expect(result.uptime).toBeDefined();
      expect(result.environment).toBeDefined();
    });

    it('should have valid timestamp in ISO format', () => {
      const result = appController.getHealthCheck();
      expect(() => new Date(result.timestamp)).not.toThrow();
    });

    it('should have positive uptime', () => {
      const result = appController.getHealthCheck();
      expect(result.uptime).toBeGreaterThan(0);
    });
  });
});
