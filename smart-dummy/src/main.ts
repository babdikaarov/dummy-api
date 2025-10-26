import { NestFactory } from '@nestjs/core';
import { SwaggerModule, DocumentBuilder } from '@nestjs/swagger';
import { AppModule } from './app.module';
import { initializeDatabase } from './database/database';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  // Initialize database
  await initializeDatabase();

  // Configure CORS
  app.enableCors({
    origin: process.env.OLOLO_MOBILE_GATE_API_ORIGIN || '*',
    credentials: true,
  });

  // Swagger configuration
  const config = new DocumentBuilder()
    .setTitle('Gates API')
    .setDescription(
      'Access control system API for managing locations and gates. ' +
        'Supports retrieving locations with gates, assigning users to locations, ' +
        'and controlling gate access.',
    )
    .setVersion('1.0')
    .addServer('http://localhost:3000', 'Development')
    .addApiKey(
      {
        type: 'apiKey',
        name: 'origin',
        in: 'header',
        description:
          'Origin header must match OLOLO_MOBILE_GATE_API_ORIGIN environment variable',
      },
      'origin',
    )
    .build();

  const document = SwaggerModule.createDocument(app, config);
  SwaggerModule.setup('api/docs', app, document, {
    jsonDocumentUrl: 'api-json',
  });

  const port = process.env.PORT ?? 3000;
  await app.listen(port);
  console.log(`Application is running on: http://localhost:${port}`);
  console.log(`Swagger documentation: http://localhost:${port}/api/docs`);
  console.log(`API JSON: http://localhost:${port}/api-json`);
}
bootstrap();
