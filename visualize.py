import socket
import pygame
import sys
from rich.console import Console
from rich.panel import Panel

console = Console()

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

console.print(Panel.fit("[bold cyan][SYSTEM INITIALIZATION][/bold cyan]\nPreparing radar canvas and display buffers...", title="✨ Radar Setup"))
console.print("🌐 [bold cyan][NETWORK][/bold cyan] Attempting to reach Go backend server at localhost:8800...")

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
try:
    client.connect(("localhost", 8800))
    client.setblocking(True)
    
    client.sendall(b"I|AM|VISUALIZATION\n")
    console.print("🤝 [bold blue][HANDSHAKE][/bold blue] Identity transmitted successfully: [magenta]I|AM|VISUALIZATION[/magenta]")
    console.print("✅ [bold green][CONNECTED][/bold green] Established stable telemetry line with Go backend.")
except Exception as e:
    console.print(Panel(f"[bold red]Critical connection error occurred![/bold red]\nCould not connect to remote host: [yellow]{e}[/yellow]", title="❌ CONNECTION FAILED"))
    sys.exit()

running = True
trajectories_dict = {}
incoming_generation = {} 
target_phys_x = None
target_phys_y = None

pygame.font.init()
font = pygame.font.SysFont("Courier", 16)

socket_file = client.makefile('r', encoding='utf-8')
try:
    console.print("📡 [bold cyan][LISTENING][/bold cyan] Awaiting target coordinate allocation matrix data...")
    raw_line = socket_file.readline().strip()
    if "TARGET|" in raw_line:
        parts = raw_line.split("|")[-1]
        x_str, y_str = parts.split(",")
        target_phys_x = float(x_str)
        target_phys_y = float(y_str)
        client.sendall(b"READY\n")
        console.print(f"🎯 [bold green][TARGET LOCKED][/bold green] Coordinates synchronized: [yellow]X={target_phys_x}m, Y={target_phys_y}m[/yellow]")
except Exception as e:
    console.print(Panel(f"Protocol mismatch or broken stream packet during baseline synchronization:\n[yellow]{e}[/yellow]", title="💥 PRIMARY HANDSHAKE ERROR"))

client.setblocking(False)
network_buffer = ""
backend_active = True

console.print("🚀 [bold green][RADAR READY][/bold green] Radar visualization matrix initialized. Streaming frames...")
while running:
    screen.fill((10, 10, 15))
    
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            console.print("\n🛑 [bold yellow][USER EXIT][/bold yellow] Shutdown requested via GUI window close event.")
            running = False

    if backend_active:
        try:
            data = client.recv(65536).decode('utf-8')
            if data == "":
                backend_active = False
                if len(incoming_generation) > 0:
                    trajectories_dict = incoming_generation.copy()
                console.print("\n✨ [bold green][SIMULATION COMPLETED][/bold green] Go backend finished processing. Final frame locked on target.")
            else:
                network_buffer += data
        except BlockingIOError:
            pass
        except (ConnectionResetError, BrokenPipeError):
            backend_active = False
            if len(incoming_generation) > 0:
                trajectories_dict = incoming_generation.copy()
            console.print("\n⚠️  [bold yellow][NETWORK WARNING][/bold yellow] Connection reset by backend. Displaying final accumulated tracks.")
        except Exception as e:
            console.print(f"\n❌ [bold red][NETWORK ERROR][/bold red] Unexpected packet stream exception: {e}")

    while "\n" in network_buffer:
        line, network_buffer = network_buffer.split("\n", 1)
        line = line.strip()
        
        if line == "FINISHWORKING":
            backend_active = False
            console.print("\n🛑 [bold yellow][SHUTDOWN][/bold yellow] Received termination command from server. Exiting...")
            break
            
        if not line or ";" not in line:
            continue
            
        parts = line.split(";")
        rocket_id = None
        phys_points = []
        
        for part in parts:
            if "ROCKET_ID:" in part:
                try:
                    rocket_id = int(part.split(":")[-1])
                except ValueError:
                    continue
            elif "," in part:
                try:
                    x_str, y_str = part.split(",")
                    phys_points.append((float(x_str), float(y_str)))
                except ValueError:
                    continue
        
        if rocket_id is not None and len(phys_points) > 0:
            color = COLORS[(rocket_id - 1) % len(COLORS)]
            incoming_generation[rocket_id] = (color, phys_points)

    if len(incoming_generation) >= 10:
        trajectories_dict = incoming_generation.copy()
        incoming_generation.clear()
        console.print(f"🧬 [bold blue][RADAR UPDATE][/bold blue] Swapping generation buffer. Rendered [cyan]{len(trajectories_dict)}[/cyan] active specimens.")
        if backend_active:
            try:
                client.sendall(b"NEXT_GEN_READY\n")
            except socket.error:
                backend_active = False

    if len(trajectories_dict) == 0 and len(incoming_generation) > 0:
        trajectories_dict = incoming_generation.copy()

    all_current_points = list(trajectories_dict.values())
    max_phys_x = target_phys_x if (target_phys_x and target_phys_x > 0) else 1.0
    max_phys_y = target_phys_y if (target_phys_y and target_phys_y > 0) else 1.0
    
    for color, phys_points in all_current_points:
        for px, py in phys_points:
            if px > max_phys_x: max_phys_x = px
            if py > max_phys_y: max_phys_y = py
                
    scale_x = (WIDTH - 160) / max_phys_x if max_phys_x > 0 else 1
    scale_y = (HEIGHT - 160) / max_phys_y if max_phys_y > 0 else 1
    final_scale = min(scale_x, scale_y) if max_phys_x > 0 and max_phys_y > 0 else 0.1
    
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

    status_text = "RADAR TELEMETRY STREAMING LIVE" if backend_active else "SIMULATION OVER: TARGET EVALUATED"
    status_color = (0, 255, 200) if backend_active else (255, 255, 50)
    
    timer_surface = font.render(status_text, True, status_color)
    count_surface = font.render(f"RENDERED TRACKS: {len(trajectories_dict)}", True, (255, 255, 255))
    screen.blit(timer_surface, (20, 20))
    screen.blit(count_surface, (20, 45))

    pygame.display.flip()
    clock.tick(60)

console.print(Panel.fit("[bold green]System resource teardown initiated successfully.[/bold green]\nReleasing socket handles and shutting down Pygame window.", title="🔌 SHUTDOWN"))
socket_file.close()
client.close()
pygame.quit()
sys.exit()