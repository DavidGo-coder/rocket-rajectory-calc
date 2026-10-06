import socket
import pygame
import sys
import time

WIDTH, HEIGHT = 800, 600
pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Live Rocket Telemetry Radar")
clock = pygame.time.Clock()

COLORS = [
    (255, 50, 50), (50, 255, 50), (50, 50, 255), (255, 255, 50), 
    (255, 50, 255), (0, 255, 255), (255, 128, 0), (128, 0, 128), 
    (0, 128, 128), (128, 128, 128)
]

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
try:
    client.connect(("localhost", 8800))
    client.setblocking(True)
    print("Успешное подключение к бэкенду Go.")
except Exception as e:
    print("Ошибка подключения:", e)
    sys.exit()

# Оборачиваем сокет в файл для удобного чтения построчно через readline()
socket_file = client.makefile('r', encoding='utf-8')

running = True
trajectories_dict = {}
incoming_generation = {} 
last_generation_swap_time = time.time()
GENERATION_DISPLAY_DURATION = 3.0

target_phys_x = None
target_phys_y = None

pygame.font.init()
font = pygame.font.SysFont("Courier", 16)

# Первичное чтение пакета цели при запуске
try:
    raw_line = socket_file.readline().strip()
    if "TARGET|" in raw_line:
        parts = raw_line.split("|")[1]
        x_str, y_str = parts.split(",")
        target_phys_x = float(x_str)
        target_phys_y = float(y_str)
        client.sendall(b"READY\n")
except Exception as e:
    print("Ошибка первичного хэндшейка:", e)

while running:
    screen.fill((10, 10, 15))
    current_time = time.time()
    
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False

    # Читаем данные от Go, только если буфер пуст и пришло время обновлять поколение
    if len(incoming_generation) == 0:
        incoming_generation.clear()
        for _ in range(10):
            try:
                raw_line = socket_file.readline()
                if not raw_line:
                    break
                raw_line = raw_line.strip()
                
                if ";" not in raw_line:
                    continue
                    
                parts = raw_line.split(";")
                coord_pairs = []
                rocket_id = 0
                
                for part in parts:
                    if "ROCKET_ID:" in part:
                        rocket_id = int(part.split(":")[-1])
                    elif "," in part:
                        coord_pairs.append(part)
                
                phys_points = []
                for pair in coord_pairs:
                    x_str, y_str = pair.split(",")
                    phys_points.append((float(x_str), float(y_str)))
                
                if len(phys_points) > 0:
                    color = COLORS[(rocket_id - 1) % len(COLORS)]
                    incoming_generation[rocket_id] = (color, phys_points)
            except Exception as e:
                pass
        
        if len(trajectories_dict) == 0:
            trajectories_dict = incoming_generation.copy()

    if current_time - last_generation_swap_time >= GENERATION_DISPLAY_DURATION:
        if len(incoming_generation) > 0:
            trajectories_dict = incoming_generation.copy()
            incoming_generation.clear() # Освобождаем буфер для следующей итерации
            last_generation_swap_time = current_time
            try:
                client.sendall(b"NEXT_GEN_READY\n") # Пингуем Гошу, что готовы принять новые данные
            except socket.error:
                pass

    all_current_points = list(trajectories_dict.values())
    max_phys_x = target_phys_x if (target_phys_x and target_phys_x > 0) else 1.0
    max_phys_y = target_phys_y if (target_phys_y and target_phys_y > 0) else 1.0
    
    for color, phys_points in all_current_points:
        for px, py in phys_points:
            if px > max_phys_x: max_phys_x = px
            if py > max_phys_y: max_phys_y = py
                
    scale_x = (WIDTH - 160) / max_phys_x if max_phys_x > 0 else 1
    scale_y = (HEIGHT - 160) / max_phys_y if max_phys_y > 0 else 1
    final_scale = min(scale_x, scale_y)
    
    if final_scale > 10000 or final_scale < 0.0001: 
        final_scale = 0.1

    if target_phys_x is not None and target_phys_y is not None:
        target_screen_x = int(target_phys_x * final_scale) + 80
        target_screen_y = HEIGHT - (int(target_phys_y * final_scale) + 80)
        pygame.draw.circle(screen, (255, 50, 50), (target_screen_x, target_screen_y), 12, 2)
        pygame.draw.circle(screen, (255, 50, 50), (target_screen_x, target_screen_y), 2)
        pygame.draw.line(screen, (255, 50, 50), (target_screen_x - 18, target_screen_y), (target_screen_x + 18, target_screen_y), 1)
        pygame.draw.line(screen, (255, 50, 50), (target_screen_x, target_screen_y - 18), (target_screen_x, target_screen_y + 18), 1)

    for color, phys_points in trajectories_dict.values():
        screen_points = []
        for px, py in phys_points:
            screen_x = int(px * final_scale) + 80
            screen_y = HEIGHT - (int(py * final_scale) + 80)
            screen_points.append((screen_x, screen_y))
        
        if len(screen_points) > 1:
            pygame.draw.lines(screen, color, False, screen_points, 2)
        elif len(screen_points) == 1:
            pygame.draw.circle(screen, color, screen_points[0], 3)

    seconds_left = max(0.0, GENERATION_DISPLAY_DURATION - (current_time - last_generation_swap_time))
    timer_surface = font.render(f"NEXT GEN SWAP IN: {seconds_left:.1f}s", True, (0, 255, 200))
    count_surface = font.render(f"ACTIVE TRACKS: {len(trajectories_dict)}", True, (255, 255, 255))
    screen.blit(timer_surface, (20, 20))
    screen.blit(count_surface, (20, 45))

    pygame.display.flip()
    clock.tick(60)

socket_file.close()
client.close()
pygame.quit()
sys.exit()