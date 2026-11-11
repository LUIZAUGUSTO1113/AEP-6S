try {
  rs.status();
} catch (error) {
  if (error.code !== 94) {
    throw error;
  }
  rs.initiate({
    _id: "rs0",
    members: [{ _id: 0, host: process.env.MONGO_REPLICA_MEMBER }],
  });
}

quit(db.hello().isWritablePrimary ? 0 : 1);
