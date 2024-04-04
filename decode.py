# # !/usr/bin/env python3

# from vosk import Model, KaldiRecognizer, SetLogLevel
# import sys
# import os
# import wave

# SetLogLevel(0)

# if len(sys.argv) == 2:
#     wf = wave.open(sys.argv[1], "rb")
# else:
#     wf = wave.open("decoder-test.wav", "rb")

# if wf.getnchannels() != 1 or wf.getsampwidth() != 2 or wf.getcomptype() != "NONE":
#     print ("Audio file must be WAV format mono PCM.")
#     exit (1)

# model = Model(".")
# rec = KaldiRecognizer(model, wf.getframerate())
# rec.SetWords(True)

# while True:
#     data = wf.readframes(4000)
#     if len(data) == 0:
#         break
#     if rec.AcceptWaveform(data):
#         print(rec.Result())
#     else:
#         print(rec.PartialResult())

# print(rec.FinalResult())

# import io
# from vosk import Model, KaldiRecognizer
# import sys
# import json
# import os
# import time
# import wave
# import struct

# model = Model(r"./vosk-model-ru-0.42") # Model(r"/home/user/vosk-model-ru-0.10")

# wf = wave.open(sys.argv[1], "rb") # wave.open('downloaded_files\\' + r'decoder-test.wav', "rb")
# rec = KaldiRecognizer(model, wf.getframerate())

# result = ''
# last_n = False

# while True:
#     data = wf.readframes(4000) # wf.getframerate()
#     if len(data) == 0:
#         break

#     if rec.AcceptWaveform(data):
#         res = json.loads(rec.Result())

#         if res['text'] != '':
#             result += f" {res['text']}"
#             last_n = False
#         # elif not last_n:
#         #     result += '\n'
#         #     last_n = True
#         else:
#             wf.close()
#             sys.exit(1)

# res = json.loads(rec.FinalResult())
# result += f" {res['text']}"

# wf.close()

# # перевод в массив рун
# char_set = list(map(ord, result))

# # вывод результата
# for value in char_set[:-1]:
#     print(value, end=' ')
# print(char_set[-1], end='')

# sys.exit(0)

from vosk import Model, KaldiRecognizer, SetLogLevel
from pydub import AudioSegment
import subprocess
import json
import os
import sys
import pathlib
# import torch

file_name = pathlib.Path(sys.argv[1]).name.split('.mp3')[0] + '.txt'
result_dir = pathlib.Path(sys.argv[2])
chose_model_lang = ''
if sys.argv[3] == 'ru':
    chose_model_lang = r"vosk-model-ru-0.22" #  r"vosk-model-small-ru-0.22" r"./vosk-model-ru-0.42"
elif sys.argv[3] == 'en':
    chose_model_lang = r"./vosk-model-en-us-0.22" # r"vosk-model-en-us-0.22-lgraph"

SetLogLevel(0)

# # Проверяем наличие модели
# if not os.path.exists("model"):
#     print ("Please download the model from https://alphacephei.com/vosk/models and unpack as 'model' in the current folder.")
#     exit (1)

# Устанавливаем Frame Rate
FRAME_RATE = 16000
CHANNELS=1

model = Model(chose_model_lang)
rec = KaldiRecognizer(model, FRAME_RATE)
rec.SetWords(True)

# Используя библиотеку pydub делаем предобработку аудио
mp3 = AudioSegment.from_mp3(sys.argv[1])
mp3 = mp3.set_channels(CHANNELS)
mp3 = mp3.set_frame_rate(FRAME_RATE)

# Преобразуем вывод в json
rec.AcceptWaveform(mp3.raw_data)
res = json.loads(rec.FinalResult())

result1 = f"{res['text']}"
result2 = res['result']
result2 = [[debug_value['start'], debug_value['word']] for debug_value in result2]
final_parsed_string_right = ""
for i in result2:
    final_parsed_string_right += str(i[0]) + "," + i[1] + ","



# # Добавляем пунктуацию

#model, example_texts, languages, punct, 
# _, _, _, _, apply_te = torch.hub.load(repo_or_dir='snakers4/silero-models',
#                                                                   model='silero_te')
# result1 = apply_te(result1)

f = open(os.path.join(pathlib.Path.cwd(), result_dir, file_name), 'w', encoding='utf-8')
f.write(result1 + '|')
f.write(final_parsed_string_right[:-1])
f.close()

# print(result1)

os._exit(0)

result_char_set = list(map(ord, result))
output = ""
# вывод результата
for value in result_char_set[:-1]:
    # print(value, end=' ')
    output += str(value) + ' '
# print(result_char_set[-1], end='')
output += str(result_char_set[-1]) + ' '

# with open(result_dir + file_name, 'w') as f:
    # json.dump(result, f, ensure_ascii=False, indent=4)
f = open(result_dir + file_name, 'w')
f.write(output)
f.close()

# перевод в массив рун вывод скрипта, для передачи в go
file_name_char_set = list(map(ord, file_name))

# вывод результата
for value in file_name_char_set[:-1]:
    print(value, end=' ')
print(file_name_char_set[-1], end='')