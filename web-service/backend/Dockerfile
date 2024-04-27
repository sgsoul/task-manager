FROM python:3.9

COPY requirements.txt requirements.txt
RUN python -m pip install --upgrade pip && pip install -r requirements.txt

WORKDIR /app

COPY ./wsgi.py wsgi.py
COPY ./app.py app.py

EXPOSE 8000

ENTRYPOINT ["gunicorn", "--bind", "0.0.0.0:8000", "wsgi:app"]