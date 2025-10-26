import { db } from './database';
import { locations, gates } from './schema';

// Sample location names and addresses
const locationNames = [
  'Торгово-развлекательный центр Ала-Тоо',
  'Гостинично-развлекательный комплекс Конуср',
  'Автопарк "Бишкек Авто"',
  'Офисный центр "Мерибель"',
  'Жилой комплекс "Белый город"',
  'Торговый центр "Дордой"',
  'Производственный парк "Текнопарк"',
  'Спортивный комплекс "Динамо"',
  'Логистический центр "Шелковый путь"',
  'Бизнес-центр "Орто Сай"',
  'Паркинг "Центральный"',
  'Гостиница "Тиан Шань"',
  'Кемпинг "Горная база"',
  'Склад оптовой торговли "Мега"',
  'Автосервис "Мастер"',
  'Ресторан "Чайхана"',
  'Морг-доставка "Доставляй"',
  'Библиотека "Абая Кунанбаева"',
  'Парк культуры имени Панфилова',
  'Музей истории Киргизии',
];

const addresses = [
  'г. Бишкек, проспект Чуй, 135',
  'г. Бишкек, улица Московская, 89',
  'г. Бишкек, проспект Манаса, 56',
  'г. Бишкек, улица Киевская, 234',
  'г. Бишкек, проспект Правобережный, 142',
  'г. Бишкек, улица Ленина, 78',
  'г. Бишкек, улица Тынстанова, 45',
  'г. Бишкек, улица Закамиля, 67',
  'г. Бишкек, проспект Булвар, 112',
  'г. Бишкек, улица Гоголя, 89',
  'г. Бишкек, улица Осокина, 123',
  'г. Бишкек, улица Союзная, 156',
  'г. Бишкек, улица Жораева, 34',
  'г. Бишкек, проспект Абдрахманова, 78',
  'г. Бишкек, улица Юнусалиева, 56',
  'г. Бишкек, улица Ганди, 45',
  'г. Бишкек, улица Логинова, 23',
  'г. Бишкек, улица Боконбаева, 89',
  'г. Бишкек, улица Холодова, 12',
  'г. Бишкек, улица Анкара, 90',
];

const gateDescriptions = [
  'Main vehicle entrance for visitors. Controlled by biometric access, opens in 3 seconds with safety sensors.',
  'Emergency exit gate with manual override. Accessible during non-business hours.',
  'Loading dock gate for delivery vehicles. Restricted access, requires special authorization.',
  'Pedestrian entrance gate. Automatic sliding doors, equipped with proximity sensors.',
  'Service vehicle entrance. Monitored 24/7 by security cameras.',
  'VIP parking entrance. Premium access control with facial recognition.',
  'Secondary exit for personnel. Restricted access with RFID card readers.',
  'Fire escape gate. Emergency exit only, alarmed against unauthorized opening.',
  'Loading/Unloading area. Industrial-grade barrier gate for heavy vehicles.',
  'Temporary access gate. Seasonal entrance, secured with bollards.',
  'Maintenance entrance. Technical staff only, requires keycard activation.',
  'Visitor parking entrance. Ticketed access system with automatic counters.',
];

const gateNames = [
  'Автоматический Шлагбаум №1',
  'Ворота входа №2',
  'Боковой въезд №3',
  'Служебный вход №4',
  'Пешеходные ворота №5',
  'Аварийный выход №6',
  'Грузовой выезд №7',
  'Ворота парковки №8',
  'Входная секция №9',
  'Выездная рампа №10',
  'Контрольный пункт №11',
  'Техническая дверь №12',
];

function getRandomItem<T>(array: T[]): T {
  return array[Math.floor(Math.random() * array.length)];
}

function getRandomNumber(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

async function seed() {
  try {
    console.log('Starting database seed...');

    // Create locations data
    const locationsData = locationNames.map((name, index) => ({
      id: index + 1,
      title: name,
      address: addresses[index],
      logo: `https://picsum.photos/seed/location${index + 1}/200`,
    }));

    // Insert locations
    console.log(`Inserting ${locationsData.length} locations...`);
    await db.insert(locations).values(locationsData).onConflictDoNothing();

    // Create gates data
    const gatesData: (typeof gates.$inferInsert)[] = [];
    let gateId = 1;

    for (const location of locationsData) {
      // Random number of gates per location (1-3)
      const gateCount = getRandomNumber(1, 3);

      for (let i = 0; i < gateCount; i++) {
        gatesData.push({
          id: gateId++,
          title: getRandomItem(gateNames),
          description: getRandomItem(gateDescriptions),
          locationId: location.id,
          isOpen: Math.random() > 0.5,
          gateIsHorizontal: Math.random() > 0.5,
        });
      }
    }

    // Insert gates
    console.log(`Inserting ${gatesData.length} gates...`);
    await db.insert(gates).values(gatesData).onConflictDoNothing();

    console.log('\n✅ Seed completed successfully!');
    console.log(`   📍 Total locations: ${locationsData.length}`);
    console.log(`   🚪 Total gates: ${gatesData.length}`);
    console.log(
      `   📊 Average gates per location: ${(gatesData.length / locationsData.length).toFixed(2)}`,
    );
  } catch (error) {
    console.error('❌ Seed failed:', error);
    throw error;
  }
}

seed();
