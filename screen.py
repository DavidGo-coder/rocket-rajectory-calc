import socket
import pygame
import sys
import threading
import time
import math
from rich.console import Console
from rich.panel import Panel

console = Console()

# --- ИНИЦИАЛИЗАЦИЯ И ПОЛНОЭКРАННЫЙ РЕЖИМ ---
pygame.init()
pygame.font.init()

info_object = pygame.display.Info()
WIDTH, HEIGHT = info_object.current_w, info_object.current_h

screen = pygame.display.set_mode((WIDTH, HEIGHT), pygame.FULLSCREEN)
pygame.display.set_caption("🛸 Core Global 3D Mission Control")
clock = pygame.time.Clock()
font = pygame.font.SysFont("Courier", 16)
font_large = pygame.font.SysFont("Courier", 22, bold=True)

# --- ГЕОМЕТРИЯ ЭКРАНА ---
TOP_PANEL_HEIGHT = int(HEIGHT * 0.70)
RADAR_WIDTH = WIDTH // 2
CHART_WIDTH = WIDTH // 2
BOTTOM_PANEL_START_Y = TOP_PANEL_HEIGHT
BOTTOM_PANEL_HEIGHT = HEIGHT - TOP_PANEL_HEIGHT

# Строгая и чистая цветовая палитра
BG_COLOR = (12, 12, 18)
GRID_COLOR = (25, 25, 35)
TEXT_COLOR = (240, 240, 245)
LINE_COLOR = (0, 255, 150)
CHAMP_COLOR = (255, 215, 0)
PARSER_BG = (18, 18, 26)
COLORS = [(255, 50, 50), (50, 255, 50), (50, 50, 255), (255, 255, 50), (255, 50, 255), (0, 255, 255)]

# --- ПАРАМЕТРЫ ИНТЕРАКТИВНОЙ ТРЕХМЕРНОЙ КАМЕРЫ ---
cam_pitch = -0.4   # Угол наклона камеры вверх/вниз
cam_yaw = 0.6      # Угол поворота сцены влево/вправо
cam_zoom = 1.2     # Базовый масштаб проекции
is_dragging = False
last_mouse_pos = (0, 0)

# --- ГЛОБАЛЬНЫЕ ДАННЫЕ И МЬЮТЕКС ЗАЩИТЫ ---
data_lock = threading.Lock()

system_running = True
radar_active = True
printer_active = True

# Трехмерные физические координаты
target_phys_x, target_phys_y, target_phys_z = None, None, None
trajectories_dict = {}
incoming_generation = {}

# Данные Принтера
max_generations = 100
median_history = []

# ТТХ лучшей ракеты на панели GUI
champion_stats = {
    "Pitch": "0.00", "Yaw": "0.00", "Fuel": "0.00",
    "BurnRate": "0.00", "Mass": "0.00", "MissDistance": "N/A",
    "IsFinal": False
}

# Интерактивный Парсер
input_fields = {
    "Target X": "5000", "Target Y": "30000", "Target Z": "15000",
    "PopSize": "100", "Gens": "30", "Pitch": "55.0",
    "Yaw": "35.0", "Fuel": "30.0", "BurnRate": "5.0",
    "BodyMass": "15.0"
}
active_field = None
field_rects = {}
status_message = "READY TO INJECT CONFIG"

trigger_injection = False
injection_data = ""

# --- МАТЕМАТИЧЕСКАЯ ФУНКЦИЯ ПРОЕКЦИИ 3D -> 2D ---
def project_3d(x, y, z, max_bound):
    global cam_pitch, cam_yaw, cam_zoom
    
    if max_bound is None or math.isnan(max_bound) or max_bound <= 0.0: norm = 1.0
    else: norm = float(max_bound)
        
    x = 0.0 if math.isnan(x) or math.isinf(x) else x
    y = 0.0 if math.isnan(y) or math.isinf(y) else y
    z = 0.0 if math.isnan(z) or math.isinf(z) else z
    
    # ИСПРАВЛЕНО: Выравниваем горизонтальную плоскость (X, Z) и высоту (Y)
    nx = (x / norm) * 260 - 130
    nz = (z / norm) * 260 - 130
    ny = (y / norm) * 260 - 130 
    
    # Вращаем горизонтальный круг (X-Z) вокруг вертикальной оси высоты Y
    x1 = nx * math.cos(cam_yaw) - nz * math.sin(cam_yaw)
    z1 = nx * math.sin(cam_yaw) + nz * math.cos(cam_yaw)
    
    # Наклоняем камеру (Pitch) относительно высоты
    y2 = ny * math.cos(cam_pitch) - z1 * math.sin(cam_pitch)
    
    screen_x = int(RADAR_WIDTH // 2 + x1 * cam_zoom)
    screen_y = int(TOP_PANEL_HEIGHT // 2 - y2 * cam_zoom)
    return screen_x, screen_y


# --- ПОТОК 1: RADAR DISPATCHER (Слушает ROCKET_ID|id|X,Y,Z;...) ---
def radar_worker():
    global target_phys_x, target_phys_y, target_phys_z, trajectories_dict, incoming_generation, radar_active
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        s.connect(("localhost", 8800))
        s.sendall(b"I|AM|VISUALIZATION\n")
        s.setblocking(False)
        buffer = ""
        target_synced = False
        
        while system_running and radar_active:
            try:
                data = s.recv(65536).decode('utf-8')
                if not data: break
                buffer += data
            except BlockingIOError:
                time.sleep(0.01)
                continue
            except Exception: break
                
            while "\n" in buffer:
                line, buffer = buffer.split("\n", 1)
                line = line.strip()
                if line == "FINISHWORKING":
                    radar_active = False
                    break
                
                # Парсинг 3D координат цели: TARGET|X|Y|Z
                if line.startswith("TARGET|") and not target_synced:
                    parts = line.split("|")
                    with data_lock:
                        target_phys_x = float(parts[1])
                        target_phys_y = float(parts[2])
                        target_phys_z = float(parts[3]) if len(parts) > 3 else 0.0
                    s.setblocking(True)
                    s.sendall(b"READY\n")
                    s.setblocking(False)
                    target_synced = True
                    continue

                    
                # ЗОЛОТОЙ СТАНДАРТ: Чтение трехмерного массива точек траекторий особей
                if line.startswith("ROCKET_ID|"):
                    parts = line.split("|")
                    try:
                        rocket_id = int(parts[1])
                        csv_points = parts[2]
                        phys_points = []
                        for pt in csv_points.split(";"):
                            if "," in pt:
                                coords = pt.split(",")
                                px = float(coords[0])
                                py = float(coords[1])
                                pz = float(coords[2]) if len(coords) > 2 else 0.0
                                phys_points.append((px, py, pz))
                        
                        if phys_points:
                            with data_lock:
                                incoming_generation[rocket_id] = (COLORS[(rocket_id - 1) % len(COLORS)], phys_points)
                    except (ValueError, IndexError):
                        continue

                # ИСПРАВЛЕНО: Синхронизация по сигналу конца поколения от Go
                if line == "END_OF_GENERATION":
                    with data_lock:
                        trajectories_dict = incoming_generation.copy()
                        incoming_generation.clear()
                    s.setblocking(True)
                    s.sendall(b"NEXT_GEN_READY\n")
                    s.setblocking(False)

    except Exception: pass
    finally: s.close()

# --- ПОТОК 2: METRICS LOG PRINTER (CHARACTERISTICS|gens, MEDIAN|val, CHAMPION) ---
def printer_worker():
    global max_generations, median_history, printer_active, champion_stats
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        s.connect(("localhost", 8800))
        s.sendall(b"I|AM|PRINTER\n")
        s.setblocking(False)
        buffer = ""
        specs_synced = False
        
        while system_running and printer_active:
            try:
                data = s.recv(65536).decode('utf-8')
                if not data: break
                buffer += data
            except BlockingIOError:
                time.sleep(0.01)
                continue
            except Exception: break
                
            while "\n" in buffer:
                line, buffer = buffer.split("\n", 1)
                line = line.strip()
                if line == "FINISHWORKING":
                    printer_active = False
                    break
                if not line or "|" not in line: continue
                    
                parts = line.split("|")
                if parts[0] == "CHARACTERISTICS" and not specs_synced:
                    with data_lock: max_generations = int(parts[1])
                    specs_synced = True
                    continue
                elif parts[0] == "MEDIAN":
                    median_val = float(parts[1])
                    with data_lock: median_history.append(median_val)
                    bar = "█" * max(1, min(20, int(500 / (median_val + 1)))) + "░" * (20 - max(1, min(20, int(500 / (median_val + 1)))))
                    console.print(f"📉 [cyan][GLOBAL LOG][/cyan] Gen {len(median_history):>3}: Median {median_val:8.2f} |[green]{bar}[/green]|")
                elif parts[0] == "GENERATION_CHAMPION":
                    with data_lock:
                        champion_stats["Pitch"] = f"{float(parts[1]):.2f}°"
                        champion_stats["Yaw"] = f"{float(parts[2]):.2f}°"
                        champion_stats["Fuel"] = f"{float(parts[3]):.1f} kg"
                        champion_stats["BurnRate"] = f"{float(parts[4]):.1f} kg/s"
                        champion_stats["Mass"] = f"{float(parts[5]):.1f} kg"
                        champion_stats["MissDistance"] = f"{float(parts[9]):.2f} m"
                        champion_stats["IsFinal"] = False
                elif parts[0] == "CHAMPION":
                    with data_lock:
                        champion_stats["Pitch"] = f"{float(parts[1]):.2f}°"
                        champion_stats["Yaw"] = f"{float(parts[2]):.2f}°"
                        champion_stats["Fuel"] = f"{float(parts[3]):.1f} kg"
                        champion_stats["BurnRate"] = f"{float(parts[4]):.1f} kg/s"
                        champion_stats["Mass"] = f"{float(parts[5]):.1f} kg"
                        champion_stats["MissDistance"] = f"{float(parts[9]):.2f} m"
                        champion_stats["IsFinal"] = True
                    printer_active = False
    except Exception: pass
    finally: s.close()

# --- ПОТОК 3: CONFIG INJECTOR (Инжектирует строго палки) ---
def parser_worker():
    global trigger_injection, status_message
    while system_running:
        if trigger_injection:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            try:
                s.connect(("localhost", 8800))
                s.sendall(b"I|AM|PARSER\n")
                time.sleep(0.15)
                with data_lock: payload = injection_data
                s.sendall(payload.encode('utf-8'))
                with data_lock: status_message = "ENGINE RUNNING"
            except Exception as e:
                with data_lock:
                    status_message = f"ERROR: {e}"
            finally:
                s.close()
                trigger_injection = False
        time.sleep(0.05)

# --- ИНЖЕКЦИЯ И СТАРТ ВСЕХ СЕТЕВЫХ ПОТОКОВ В ФОНЕ ---
threading.Thread(target=radar_worker, daemon=True).start()
threading.Thread(target=printer_worker, daemon=True).start()
threading.Thread(target=parser_worker, daemon=True).start()

# --- ГЛАВНЫЙ ЦИКЛ ОБРАБОТКИ СОБЫТИЙ И ОТРИСОВКИ ---
while system_running:
    screen.fill(BG_COLOR)
    mx, my = pygame.mouse.get_pos()
    
    # 📋 Перехват ввода пользователя (клавиатура и клики мыши)
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            system_running = False
        if event.type == pygame.KEYDOWN:
            if event.key == pygame.K_ESCAPE:
                system_running = False
                
            # ИСПРАВЛЕНО: Зум работает ТОЛЬКО если пользователь НЕ вводит текст прямо сейчас
            elif (event.key == pygame.K_EQUALS or event.key == pygame.K_KP_PLUS) and active_field is None:
                cam_zoom = min(5.0, cam_zoom + 0.1)
            elif (event.key == pygame.K_MINUS or event.key == pygame.K_KP_MINUS) and active_field is None:
                cam_zoom = max(0.2, cam_zoom - 0.1)
                
            # Если же какое-то поле ввода активно — обрабатываем текст
            elif active_field is not None:
                if event.key == pygame.K_BACKSPACE:
                    input_fields[active_field] = input_fields[active_field][:-1]
                elif event.key == pygame.K_RETURN:
                    active_field = None
                else:
                    # Разрешаем вводить цифры, точку и коварный знак МИНУС
                    if len(input_fields[active_field]) < 12:
                        if event.unicode in "0123456789.-":
                            input_fields[active_field] += event.unicode


                        
        # Интерактивное вращение 3D-пространства мышкой
        if event.type == pygame.MOUSEBUTTONDOWN:
            if event.button == 1:
                active_field = None
                for field_name, rect in field_rects.items():
                    if rect.collidepoint(mx, my):
                        active_field = field_name
                        
                # Клик по кнопке инжекции параметров (Золотой стандарт палок)
                                # Клик по кнопке инжекции параметров с жесткой фильтрацией хлама
                                # Клик по кнопке инжекции параметров (Валидацию углов отдаем полностью в Go)
                                # Клик по кнопке инжекции параметров (Разрешаем отрицательные X и Z для 4-х квадрантов)
                if launch_btn_rect.collidepoint(mx, my) and not trigger_injection:
                    try:
                        # 1. Проверяем целые числа
                        tgt_x = int(input_fields["Target X"])
                        tgt_y = int(input_fields["Target Y"]) # Высота
                        tgt_z = int(input_fields["Target Z"])
                        pop_size = int(input_fields["PopSize"])
                        gens_count = int(input_fields["Gens"])
                        
                        # ИСПРАВЛЕНО: Убраны ограничения на знак минус для X и Z. Только Y (высота) обязана быть >= 0!
                        if tgt_y < 0 or pop_size <= 0 or gens_count <= 0:
                            raise ValueError("Height, size, and generations must be positive!")

                        # 2. Проверяем ТТХ ракеты (float)
                        pitch_val = float(input_fields["Pitch"])
                        yaw_val = float(input_fields["Yaw"])
                        fuel_val = float(input_fields["Fuel"])
                        burn_val = float(input_fields["BurnRate"])
                        mass_val = float(input_fields["BodyMass"])
                        
                        if fuel_val < 0 or burn_val <= 0 or mass_val <= 0:
                            raise ValueError("Mass, fuel, and rate must be positive!")

                        if abs(pitch_val) > 3600.0 or abs(yaw_val) > 3600.0:
                            raise ValueError("Angle values are too extreme!")

                        efficiency_val = 800.0 + (burn_val / mass_val) * 1500.0
                        
                        line1 = f"{tgt_x}|{tgt_y}|{tgt_z}\n"
                        line2 = f"{pitch_val}|{yaw_val}|{fuel_val}|{burn_val}|{mass_val}|{efficiency_val:.1f}\n"
                        line3 = f"{pop_size}|{gens_count}\n"
                        
                        injection_data = line1 + line2 + line3
                        status_message = "TRANSMITTING DATA MATRIX..."
                        trigger_injection = True
                        
                    except ValueError as e:
                        status_message = f"INPUT ERROR: {str(e) if 'must' in str(e) or 'extreme' in str(e) else 'Enter valid numbers only!'}"


                        
                # Проверяем зажатие ЛКМ в зоне 3D радара для вращения камеры
                if my < TOP_PANEL_HEIGHT and mx < RADAR_WIDTH:
                    is_dragging = True
                    last_mouse_pos = (mx, my)
                    
        if event.type == pygame.MOUSEBUTTONUP:
            if event.button == 1:
                is_dragging = False
                
        # ИСПРАВЛЕНО: Распаковка кортежа координат мыши для правильного вычитания
        if event.type == pygame.MOUSEMOTION:
            if is_dragging:
                # Извлекаем отдельные числа X и Y из сохраненной позиции
                last_mx, last_my = last_mouse_pos 
                dx = mx - last_mx
                dy = my - last_my
                
                cam_yaw += dx * 0.007
                cam_pitch = max(-math.pi/2, min(math.pi/2, cam_pitch + dy * 0.007))
                last_mouse_pos = (mx, my)


    # Безопасный экспорт данных из фоновых сетевых потоков через мьютекс
    with data_lock:
        cur_target_x, cur_target_y, cur_target_z = target_phys_x, target_phys_y, target_phys_z
        cur_trajectories = trajectories_dict.copy()
        cur_incoming = incoming_generation.copy()
        cur_medians = median_history.copy()
        cur_max_gens = max_generations
        cur_champion = champion_stats.copy()
        cur_status = status_message
    # =========================================================================
    # 1. 🛰️ ЗОНА РАДАРA (ЛЕВЫЙ ВЕРХНИЙ СЕКТОР - НАСТОЯЩЕЕ 3D)
    # =========================================================================
    pygame.draw.line(screen, GRID_COLOR, (RADAR_WIDTH, 0), (RADAR_WIDTH, TOP_PANEL_HEIGHT), 2)
    
    if len(cur_trajectories) == 0 and cur_incoming:
        cur_trajectories = cur_incoming.copy()
        
    # ИСПРАВЛЕНО: Безопасный плоский список для max(), исключающий кортежи
    coords_list = [1.0]
    if cur_target_x is not None: coords_list.append(cur_target_x)
    if cur_target_y is not None: coords_list.append(cur_target_y)
    if cur_target_z is not None: coords_list.append(cur_target_z)
    
    # ИСПРАВЛЕНО: Строгая валидация пар координат перед отправкой в pygame.draw.lines
    for color, pts in cur_trajectories.values():
        if pts:
            s_pts = []
            for px, py, pz in pts:
                res = project_3d(px, py, pz, max_bound)
                # Проверяем, что project_3d вернул честный кортеж из двух int-чисел
                if isinstance(res, tuple) and len(res) == 2:
                    if isinstance(res[0], int) and isinstance(res[1], int):
                        s_pts.append(res)
            
            # Отрисовываем, только если у нас набралось как минимум 2 валидные точки
            if len(s_pts) > 1:
                pygame.draw.lines(screen, color, False, s_pts, 2)
                
    max_bound = max(coords_list)


    # Отрисовка трехмерной координатной сетки сцены (Оси X, Y, Z)
    o_x, o_y = project_3d(0, 0, 0, max_bound)
    ax_x, ay_x = project_3d(max_bound, 0, 0, max_bound)
    ax_y, ay_y = project_3d(0, max_bound, 0, max_bound)
    ax_z, ay_z = project_3d(0, 0, max_bound, max_bound)
    
    pygame.draw.line(screen, (150, 50, 50), (o_x, o_y), (ax_x, ay_x), 1) # Axis X
    pygame.draw.line(screen, (50, 150, 50), (o_x, o_y), (ax_y, ay_y), 2) # Axis Y (Высота)
    pygame.draw.line(screen, (50, 50, 150), (o_x, o_y), (ax_z, ay_z), 1) # Axis Z

    # Отрисовка трехмерной мишени (3D Цель)
    if cur_target_x is not None and cur_target_y is not None:
        tx, ty = project_3d(cur_target_x, cur_target_y, cur_target_z, max_bound)
        pygame.draw.circle(screen, (255, 50, 50), (tx, ty), 8, 2)
        pygame.draw.circle(screen, (255, 50, 50), (tx, ty), 2)

    # Отрисовка шлейфов трехмерного полета популяции ракет
    for color, pts in cur_trajectories.values():
        s_pts = [project_3d(px, py, pz, max_bound) for px, py, pz in pts]
        if len(s_pts) > 1:
            pygame.draw.lines(screen, color, False, s_pts, 2)

    screen.blit(font.render("🛰️ LIVE 3D TELEMETRY RADAR (DRAG TO ROTATE)", True, (0, 255, 200)), (20, 20))
    screen.blit(font.render("Use [+] and [-] keys to zoom vector field", True, (130, 130, 150)), (20, 42))

    # =========================================================================
    # 2. 📊 ЗОНА МЕДИАНЫ (ПРАВЫЙ ВЕРХНИЙ СЕКТОР)
    # =========================================================================
    chart_start_x = RADAR_WIDTH + 50
    chart_end_x = WIDTH - 50
    chart_draw_width = chart_end_x - chart_start_x

    for cx in range(chart_start_x, chart_end_x, 100):
        pygame.draw.line(screen, GRID_COLOR, (cx, 80), (cx, TOP_PANEL_HEIGHT - 50), 1)
    for cy in range(80, TOP_PANEL_HEIGHT - 50, 60):
        pygame.draw.line(screen, GRID_COLOR, (chart_start_x, cy), (chart_end_x, cy), 1)

    pygame.draw.line(screen, TEXT_COLOR, (chart_start_x, TOP_PANEL_HEIGHT - 50), (chart_end_x, TOP_PANEL_HEIGHT - 50), 2)
    pygame.draw.line(screen, TEXT_COLOR, (chart_start_x, 80), (chart_start_x, TOP_PANEL_HEIGHT - 50), 2)

    if len(cur_medians) > 1:
        mx, mn = max(cur_medians), min(cur_medians)
        rng = (mx - mn) if (mx - mn) > 0 else 1.0
        x_step = chart_draw_width / (max(cur_max_gens, len(cur_medians)) - 1)
        c_points = [(chart_start_x + int(i * x_step), (TOP_PANEL_HEIGHT - 50) - int(((v - mn) / rng) * (TOP_PANEL_HEIGHT - 170)) - 15) for i, v in enumerate(cur_medians)]
        pygame.draw.lines(screen, LINE_COLOR, False, c_points, 3)

    screen.blit(font.render("📊 FITNESS TREND (MEDIAN CONVERGENCE)", True, CHAMP_COLOR), (chart_start_x - 30, 20))


    # =========================================================================
    # 3. 📥 ТРЕХЗОННЫЙ КОМАНДНЫЙ ИНТЕРФЕЙС (НИЖНЯЯ ПАНЕЛЬ КОНФИГУРАЦИИ)
    # =========================================================================
    pygame.draw.line(screen, GRID_COLOR, (0, BOTTOM_PANEL_START_Y), (WIDTH, BOTTOM_PANEL_START_Y), 3)
    pygame.draw.rect(screen, PARSER_BG, (0, BOTTOM_PANEL_START_Y, WIDTH, BOTTOM_PANEL_HEIGHT))
    
    # Расчет ширины трех независимых окон
    ZONE_WIDTH = WIDTH // 3
    pygame.draw.line(screen, GRID_COLOR, (ZONE_WIDTH, BOTTOM_PANEL_START_Y), (ZONE_WIDTH, HEIGHT), 1)
    pygame.draw.line(screen, GRID_COLOR, (ZONE_WIDTH * 2, BOTTOM_PANEL_START_Y), (ZONE_WIDTH * 2, HEIGHT), 1)

    # --- СЕКТОР А: НАСТРОЙКА ПАРАМЕТРОВ ---
    screen.blit(font.render("🧬 SYSTEM CONFIGURATION", True, (160, 160, 220)), (20, BOTTOM_PANEL_START_Y + 15))
    start_x = 20
    start_y = BOTTOM_PANEL_START_Y + 45
    col_w = ZONE_WIDTH // 2
    field_rects.clear()

    for idx, (field_name, value) in enumerate(input_fields.items()):
        c = idx % 2
        r = idx // 2
        x_pos = start_x + (c * col_w)
        y_pos = start_y + (r * 32)
        
        screen.blit(font.render(f"{field_name}:", True, TEXT_COLOR), (x_pos, y_pos + 3))
        input_rect = pygame.Rect(x_pos + 95, y_pos, col_w - 110, 22)
        field_rects[field_name] = input_rect
        
        box_color = LINE_COLOR if active_field == field_name else (60, 60, 80)
        pygame.draw.rect(screen, box_color, input_rect, 1)
        screen.blit(font.render(value, True, (255, 255, 255)), (input_rect.x + 5, input_rect.y + 3))

    # Кнопка пуска под сеткой полей
    launch_btn_rect = pygame.Rect(20, HEIGHT - 55, ZONE_WIDTH - 40, 38)
    btn_color = (0, 180, 100) if launch_btn_rect.collidepoint(mx, my) else (0, 120, 70)
    pygame.draw.rect(screen, btn_color, launch_btn_rect)
    pygame.draw.rect(screen, LINE_COLOR, launch_btn_rect, 1)
    
    btn_text = font.render("🔥 INJECT CONFIG & IGNITE ENGINE", True, (255, 255, 255))
    screen.blit(btn_text, (launch_btn_rect.x + (launch_btn_rect.width - btn_text.get_width()) // 2, launch_btn_rect.y + 11))

    # --- СЕКТОР Б: МОНИТОР СТАТУСА СИМУЛЯЦИИ ---
    center_x = ZONE_WIDTH + 20
    screen.blit(font.render("📡 TELEMETRY STATUS MONITOR", True, (160, 160, 220)), (center_x, BOTTOM_PANEL_START_Y + 15))
    
    log_rect = pygame.Rect(center_x, BOTTOM_PANEL_START_Y + 45, ZONE_WIDTH - 40, HEIGHT - BOTTOM_PANEL_START_Y - 75)
    pygame.draw.rect(screen, (10, 10, 14), log_rect)
    pygame.draw.rect(screen, GRID_COLOR, log_rect, 1)
    
    screen.blit(font.render(f"👉 Core Engine : {cur_status}", True, CHAMP_COLOR), (log_rect.x + 15, log_rect.y + 15))
    screen.blit(font.render(f"👉 Generation  : {len(cur_medians)} / {cur_max_gens}", True, TEXT_COLOR), (log_rect.x + 15, log_rect.y + 45))
    if cur_medians:
        screen.blit(font.render(f"👉 Current Med : {cur_medians[-1]:.4f}", True, LINE_COLOR), (log_rect.x + 15, log_rect.y + 75))
        bar_len = max(1, min(25, int(500 / (cur_medians[-1] + 1))))
        bar_str = "█" * bar_len + "░" * (25 - bar_len)
        screen.blit(font.render(f"👉 Trend       : |{bar_str}|", True, LINE_COLOR), (log_rect.x + 15, log_rect.y + 105))

    # --- СЕКТОР В: КАРТОЧКА ХАРАКТЕРИСТИК ЛИДЕРА (CHAMPION SPECS) ---
    right_x = (ZONE_WIDTH * 2) + 20
    title_color = CHAMP_COLOR if cur_champion["IsFinal"] else (160, 160, 220)
    title_str = "🏆 CHAMPION TRAJECTORY MATRIX" if cur_champion["IsFinal"] else "👑 LEADER PROFILE MATRIX"
    screen.blit(font.render(title_str, True, title_color), (right_x, BOTTOM_PANEL_START_Y + 15))
    
    card_rect = pygame.Rect(right_x, BOTTOM_PANEL_START_Y + 45, ZONE_WIDTH - 40, HEIGHT - BOTTOM_PANEL_START_Y - 75)
    card_bg = (24, 24, 16) if cur_champion["IsFinal"] else (14, 14, 22)
    pygame.draw.rect(screen, card_bg, card_rect)
    pygame.draw.rect(screen, title_color, card_rect, 1)
    
    metrics = [
        ("Miss Distance:", cur_champion["MissDistance"], (255, 100, 100) if not cur_champion["IsFinal"] else CHAMP_COLOR),
        ("Launch Pitch :", cur_champion["Pitch"], TEXT_COLOR),
        ("Launch Yaw   :", cur_champion["Yaw"], TEXT_COLOR),
        ("Fuel Weight  :", cur_champion["Fuel"], LINE_COLOR),
        ("Fuel BurnRate:", cur_champion["BurnRate"], LINE_COLOR),
        ("Body Mass    :", cur_champion["Mass"], TEXT_COLOR),
    ]
    
    for i, (label, val, val_color) in enumerate(metrics):
        screen.blit(font.render(label, True, TEXT_COLOR), (card_rect.x + 15, card_rect.y + 15 + (i * 24)))
        screen.blit(font.render(val, True, val_color), (card_rect.x + 160, card_rect.y + 15 + (i * 24)))
        
    pygame.display.flip()
    clock.tick(30)

with data_lock:
    system_running = False
pygame.quit()
sys.exit()

