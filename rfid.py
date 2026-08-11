def read_data(reader):
    id, data = reader.read()
    return id, data.strip()

def write_data(reader, data):
    reader.write(data)
    return data
